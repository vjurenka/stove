package game

import (
	"encoding/binary"
	"github.com/HearthSim/hs-proto-go/pegasus/game"
	"github.com/HearthSim/hs-proto-go/pegasus/shared"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"net"
	"runtime/debug"
	"sync"
)

type Packet struct {
	Body []byte
	ID   int
}

type session struct {
	sync.Mutex
	s              *server
	c              net.Conn
	g              *Game
	player         *GamePlayer
	isSpectator    bool
	bnetId         *shared.BnetId
	quit           chan struct{}
	out            chan *Packet
	ready          bool
	historyIndex   int
	packetHandlers map[int]func(*Packet)
}

func NewSession(s *server, c net.Conn) *session {
	log.Printf("game server: accepting connection from %s", c.RemoteAddr().String())
	res := &session{}
	res.s = s
	res.c = c
	res.quit = make(chan struct{})
	res.out = make(chan *Packet, 1)
	res.packetHandlers = map[int]func(*Packet){}
	res.registerHandler(game.Ping_ID, res.onPing)
	res.registerHandler(game.Handshake_ID, res.onHandshake)
	res.registerHandler(game.SpectatorHandshake_ID, res.onSpectatorHandshake)
	return res
}

func (s *session) registerHandler(x interface{}, handler func(*Packet)) {
	id := packetIDFromProto(x)
	s.packetHandlers[id] = handler
}

func (s *session) onPing(p *Packet) {
	s.writePacket(game.Pong_ID, []byte{})
}

func (s *session) onHandshake(p *Packet) {
	h := game.Handshake{}
	err := proto.Unmarshal(p.Body, &h)
	if err != nil {
		panic(err)
	}
	game := s.s.GameFromHandle(*h.GameHandle)
	player := game.PlayerFromHandle(*h.ClientHandle)
	if player != nil {
		if *h.Password == player.Password {
			s.g = game
			s.setPlayer(player)
			s.gameSetup()
		} else {
			log.Panicf("handshake: password mismatch for %d:%d",
				*h.GameHandle, *h.ClientHandle)
		}
	} else {
		log.Panicf("handshake: player %d:%d not found",
			*h.GameHandle, *h.ClientHandle)
	}
}

func (s *session) onSpectatorHandshake(p *Packet) {
	s.isSpectator = true
	h := game.SpectatorHandshake{}
	err := proto.Unmarshal(p.Body, &h)
	if err != nil {
		panic(err)
	}
	game := s.s.GameFromHandle(*h.GameHandle)
	if *h.Password == game.SpectatorPassword {
		s.g = game
		s.setSpectator(h.GameAccountId)
	} else {
		log.Panicf("handshake: spectator password mismatch for account=%s",
			h.GameAccountId.String())
	}
}

func (s *session) setPlayer(player *GamePlayer) {
	s.g.Lock()
	defer s.g.Unlock()
	s.g.clients = append(s.g.clients, s)
	s.player = player
	s.bnetId = player.GameAccountId

	s.registerHandler(game.GetGameState_ID, s.onGetGameState)
	s.registerHandler(game.UserUI_ID, s.onUserUI)
	s.registerHandler(game.ChooseOption_ID, s.onChooseOption)
	s.registerHandler(game.ChooseEntities_ID, s.onChooseEntities)
	// TODO: mulliganing and some other stuff
}

func (s *session) setSpectator(id *shared.BnetId) {
	s.g.Lock()
	defer s.g.Unlock()
	s.g.clients = append(s.g.clients, s)
	s.isSpectator = true
	s.bnetId = id
	panic("TODO: packet handlers for spectators")
}

func (s *session) gameSetup() {
	setup := &game.GameSetup{}
	setup.Board = proto.Int32(8)
	setup.MaxSecretsPerPlayer = proto.Int32(5)
	setup.MaxFriendlyMinionsPerPlayer = proto.Int32(7)
	setup.KeepAliveFrequency = proto.Int32(30)
	buf, err := proto.Marshal(setup)
	if err != nil {
		panic(err)
	}
	s.writePacket(game.GameSetup_ID, buf)
}

func (s *session) onGetGameState(p *Packet) {
	s.ready = true
	s.updateHistory()
}

func (s *session) updateHistory() {
	if !s.ready {
		return
	}

	// TODO: filter entities by zone/controller/reveal
	s.g.Lock()
	defer s.g.Unlock()
	hist := &game.PowerHistory{}
	hist.List = s.g.history[s.historyIndex:]
	s.historyIndex = len(s.g.history)
	buf, err := proto.Marshal(hist)
	if err != nil {
		panic(err)
	}
	s.writePacket(game.PowerHistory_ID, buf)
}

func (s *session) onUserUI(p *Packet) {
	// broadcast the packet to all other connected clients
	ui := &game.UserUI{}
	err := proto.Unmarshal(p.Body, ui)
	if err != nil {
		panic(err)
	}
	// make sure this is set as the source client:
	ui.PlayerId = proto.Int32(int32(s.player.PlayerId))
	buf, err := proto.Marshal(ui)
	if err != nil {
		panic(err)
	}
	s.g.Lock()
	defer s.g.Unlock()
	for _, c := range s.g.clients {
		if c != s {
			c.writePacket(game.UserUI_ID, buf)
		}
	}
}

func (s *session) onChooseOption(p *Packet) {
	option := &game.ChooseOption{}
	err := proto.Unmarshal(p.Body, option)
	if err != nil {
		panic(err)
	}

	s.g.ChooseOption(s.player, int(*option.Id),
		int(*option.Index), int(*option.Target),
		int(*option.SubOption), int(*option.Position))
}

func (s *session) onChooseEntities(p *Packet) {
	ce := &game.ChooseEntities{}
	err := proto.Unmarshal(p.Body, ce)
	if err != nil {
		panic(err)
	}
	es := []int{}
	for _, ei := range ce.Entities {
		es = append(es, int(ei))
	}

	s.g.ChooseEntities(s.player, int(*ce.Id), es)
}

func (s *session) Close() {
	if s.g != nil {
		s.g.Lock()
		clients := []*session{}
		for _, c := range s.g.clients {
			if c != s {
				clients = append(clients, c)
			}
		}
		s.g.clients = clients
		s.g.Unlock()
	}

	close(s.quit)
	s.c.Close()
}

func (s *session) closeOnError() {
	if err := recover(); err != nil {
		log.Printf("game server: error in session: %v\n%s",
			err, string(debug.Stack()))

		s.Close()
	}
}

var endian = binary.LittleEndian

func (s *session) handle() {
	defer s.closeOnError()

	go s.processWrites()
	lenBuf := make([]byte, 8)
	packBuf := make([]byte, 0x1000)
	for {
		_, err := io.ReadFull(s.c, lenBuf)
		if err != nil {
			panic(err)
		}
		id := int(endian.Uint32(lenBuf[0:4]))
		size := int(endian.Uint32(lenBuf[4:8]))
		if size > len(packBuf) {
			packBuf = make([]byte, size*3/2)
		}
		body := packBuf[:size]
		_, err = io.ReadFull(s.c, body)
		if err != nil {
			panic(err)
		}
		if h, ok := s.packetHandlers[id]; ok {
			h(&Packet{body, id})
		} else {
			log.Panicf("no handler for packet %d", id)
		}
	}
}

func (s *session) writePacket(x interface{}, body []byte) {
	s.out <- &Packet{body, packetIDFromProto(x)}
}

func (s *session) processWrites() {
	defer s.closeOnError()

	for {
		select {
		case p := <-s.out:
			size := len(p.Body)
			buf := make([]byte, 8+size)
			endian.PutUint32(buf[0:4], uint32(p.ID))
			endian.PutUint32(buf[4:8], uint32(size))
			copy(buf[8:], p.Body)
			_, err := s.c.Write(buf)
			if err != nil {
				panic(err)
			}
		case <-s.quit:
			return
		}
	}
}
