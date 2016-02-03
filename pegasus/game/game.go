/*
util server interacts with game server as:
 util -> game: CreateGame(* GameStartInfo)
 game -> util: <Status{Ready, []PlayerJoinInfo}
 util -> each client: PlayerJoinInfo
 any client -> game: JoinInfo{}
 game <==> client: PowerHistory, Options, GameUI, etc.
 game -> util: <Status{Complete, GameResult}

type Game contains game state, holds ptr to sessions so it can kill them,
pointer to server so it can advertise its address, holds kettle instance

type session contains conn and read/write buffers for each client, manages
reconns, holds ptr to game to forward stuff

type server holds pointers to sessions so it can kill them, manages reconns;
creates sessions directly, and handles putting them in the right Game instance
based on their submitted game/client_handle + password (AuroraHandshake)

type kettle maintains connection with the sim server for an individual game,
holds game pointer to notify of events
*/
package game

import (
	"crypto/rand"
	"fmt"
	"github.com/HearthSim/hs-proto-go/pegasus/game"
	"github.com/HearthSim/hs-proto-go/pegasus/shared"
	"github.com/HearthSim/stove/bnet"
	"github.com/golang/protobuf/proto"
	"log"
	mrand "math/rand"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

type GameStartInfo struct {
	Players []PlayerInfo
}

type PlayerInfo struct {
	DisplayName   string
	GameAccountId *shared.BnetId
	HeroCardId    string
	CardIds       []string
	Premium       []bool
}

// A game account id that signals the player is an AI.
// For some reason AIs don't have 'WTCG' in their GameAccountID
var AIGameAccountID = bnet.EntityId(0x200007A00000000, 0)

type Game struct {
	sync.Mutex

	GameId            string
	GameHandle        int32
	Players           []*GamePlayer
	SpectatorPassword string
	Address           net.TCPAddr

	// TODO: reconnection information
	Result       chan *GameResult
	HasBeenSetup bool

	quit    chan struct{}
	clients []*session // protected by mutex
	server  *server
	kettle  *KettleClient

	// all game state is protected by mutex
	history             []*game.PowerHistoryData
	currentPlayer       int
	lastOptionID        int
	lastEntityChoicesID int
}

type GamePlayer struct {
	PlayerInfo

	// Equivalent to PLAYER_ID tag value
	PlayerId int

	// Used for handshake
	ClientHandle int64
	Password     string

	// Used for history syncing
	histIndex int
}

type GameResult struct{}

func CreateGame(params *GameStartInfo) *Game {
	res := &Game{}
	mrand.Seed(time.Now().UnixNano())
	res.Players = make([]*GamePlayer, 2)
	for i := 0; i < 2; i++ {
		playerInfo := &params.Players[i]
		p := &GamePlayer{}
		p.PlayerInfo = *playerInfo
		p.PlayerId = i + 1
		p.ClientHandle = int64(mrand.Int31())
		p.Password = GenPassword()
		res.Players[i] = p
	}
	res.GameHandle = mrand.Int31()
	res.GameId = fmt.Sprintf("Test %d", res.GameHandle)
	res.SpectatorPassword = GenPassword()
	res.server = gameServer
	res.Address = *(res.server.sock.Addr().(*net.TCPAddr))
	res.server.gameHandles[res.GameHandle] = res
	res.kettle = NewKettleClient(res)

	return res
}

func (g *Game) OnTagChange(entity, tag, value int) {
	// hack for making AI end turn
	if entity == 3 {
		// CURRENT_PLAYER => 1
		if tag == 23 && value == 1 {
			go func() {
				// time.Sleep(1 * time.Second)
				g.kettle.SendOption(0, 0, -1, 0)
			}()
		}
	}
	if tag == 23 && value == 1 {
		log.Printf("--- OnTagChange --- Set current player to %d", entity-1)
		g.currentPlayer = entity - 1
	}
}

func (g *Game) OnEntityChoices(choices *game.EntityChoices) {
	g.lastEntityChoicesID++
	choices.Id = proto.Int32(int32(g.lastEntityChoicesID))
	for _, client := range g.clients {
		if client.player.PlayerId == int(*choices.PlayerId) {
			buf, err := proto.Marshal(choices)
			if err != nil {
				panic(err)
			}
			client.writePacket(game.EntityChoices_ID, buf)
		}
	}
}

func (g *Game) OnHistory(histData []*game.PowerHistoryData) {
	for _, hist := range histData {
		if hist.CreateGame != nil {
			for _, player := range hist.CreateGame.Players {
				id := *player.Id - 1
				player.GameAccountId = g.Players[id].GameAccountId
				player.CardBack = proto.Int32(26)
				e := player.Entity
				ei := int(*e.Id)
				for _, t := range e.Tags {
					g.OnTagChange(ei, int(*t.Name), int(*t.Value))
				}
			}
		}
		if hist.TagChange != nil {
			t := hist.TagChange
			g.OnTagChange(int(*t.Entity), int(*t.Tag), int(*t.Value))
		}
		if hist.FullEntity != nil {
			e := hist.FullEntity
			ei := int(*e.Entity)
			for _, t := range e.Tags {
				g.OnTagChange(ei, int(*t.Name), int(*t.Value))
			}
		}
		g.history = append(g.history, hist)
	}
	for i, client := range g.clients {
		p := g.Players[i]
		histUpdate := g.history[p.histIndex:]
		histBuf, err := proto.Marshal(&game.PowerHistory{
			List: histUpdate,
		})
		if err != nil {
			panic(err)
		}
		client.writePacket(game.PowerHistory_ID, histBuf)
		p.histIndex = len(g.history)
	}
}

func (g *Game) OnOptions(options *game.AllOptions) {
	optionsBuf, err := proto.Marshal(options)
	if err != nil {
		panic(err)
	}
	for _, client := range g.clients {
		// TODO: spectator
		log.Printf("--- OnOptions --- client.player.PlayerId=%d, g.currentPlayer=%d", client.player.PlayerId, g.currentPlayer)
		if client.player.PlayerId == g.currentPlayer {
			client.writePacket(game.AllOptions_ID, optionsBuf)
		}
	}
}

func (g *Game) ChooseOption(p *GamePlayer, id, index, target, subOption, position int) {
	g.kettle.SendOption(index, target, subOption, position)
	log.Printf("ChooseOption(%v,%d,%d,%d,%d,%d)", p, id, index, target, subOption, position)
}

func (g *Game) ChooseEntities(p *GamePlayer, id int, entities []int) {
	log.Printf("ChooseEntities(%v, %d, %v)", p, id, entities)
	g.kettle.ChooseEntities(entities)
}

func (g *Game) PlayerFromHandle(clientHandle int64) *GamePlayer {
	for idx, player := range g.Players {
		if player.ClientHandle == clientHandle {
			return g.Players[idx]
		}
	}
	return nil
}

func (g *Game) Close() {
	for _, client := range g.clients {
		client.Close()
	}
	if g.kettle != nil {
		g.kettle.Close()
	}
	g.Lock()
	defer g.Unlock()
	select {
	case <-g.quit:
	default:
		close(g.quit)
	}
}

func (g *Game) CloseOnError() {
	if err := recover(); err != nil {
		log.Printf("game server error: %v\n=== STACK TRACE ===\n%s",
			err, string(debug.Stack()))
		g.Close()
	}
}

func GenPassword() string {
	buf := make([]byte, 10)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	// want printable ascii
	for i := 0; i < 10; i++ {
		buf[i] = 0x30 + (buf[i] & 0x3f)
	}
	return string(buf)
}
