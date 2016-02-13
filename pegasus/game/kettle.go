package game

import (
	"encoding/binary"
	"encoding/json"
	"github.com/HearthSim/hs-proto-go/pegasus/game"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"net"
	"os"
	"runtime/debug"
	"strconv"
)

type KettleClient struct {
	g *Game

	write        chan []byte
	quit         chan struct{}
	conn         net.Conn
	lastOptionID int32
}

func NewKettleClient(game *Game) *KettleClient {
	res := &KettleClient{}
	res.g = game
	res.quit = make(chan struct{})
	res.write = make(chan []byte)
	res.lastOptionID = 1
	res.Connect()
	res.CreateGame()
	return res
}

func (c *KettleClient) Connect() {
	hostAddr := os.Getenv("KETTLE_ADDR")
	if len(hostAddr) == 0 {
		log.Panicf("Set KETTLE_ADDR to the address of a kettle server")
	}
	conn, err := net.Dial("tcp", hostAddr)
	if err != nil {
		panic(err)
	}
	c.conn = conn
	go c.readLoop()
	go c.writeLoop()
}

func (c *KettleClient) CreateGame() {
	init := KettleCreateGame{}
	for _, player := range c.g.Players {
		init.Players = append(init.Players, KettleCreatePlayer{
			Hero:  player.HeroCardId,
			Cards: player.CardIds,
			Name:  player.DisplayName,
		})
	}
	initBuf, err := json.Marshal([]KettlePacket{{
		Type:       "CreateGame",
		GameID:     c.g.GameId,
		CreateGame: &init,
	}})
	if err != nil {
		panic(err)
	}
	c.write <- initBuf
}

func (c *KettleClient) SendOption(index, target, subOption, position int) {
	endTurn := &KettlePacket{
		Type:   "SendOption",
		GameID: c.g.GameId,
		SendOption: &KettleSendOption{
			Index:     index,
			Target:    target,
			SubOption: subOption,
			Position:  position,
		},
	}
	endTurnBuf, _ := json.Marshal(endTurn)
	c.write <- endTurnBuf
}

func (c *KettleClient) ChooseEntities(entities []int) {
	packet := &KettlePacket{
		Type:           "ChooseEntities",
		GameID:         c.g.GameId,
		ChooseEntities: entities,
	}
	packetBuf, _ := json.Marshal(packet)
	c.write <- packetBuf
}

func (c *KettleClient) Concede(playerId int) {
	packet := &KettlePacket{
		Type:    "Concede",
		GameID:  c.g.GameId,
		Concede: playerId,
	}
	packetBuf, _ := json.Marshal(packet)
	c.write <- packetBuf
}

func (c *KettleClient) Close() {
	c.conn.Close()
}

func (c *KettleClient) closeOnError() {
	if err := recover(); err != nil {
		log.Printf("game server: error in kettle client: %v\n%s",
			err, string(debug.Stack()))

		c.Close()
		close(c.quit)
	}
}

func (c *KettleClient) readLoop() {
	defer c.closeOnError()

	lenBuf := make([]byte, 4)
	packetBuf := make([]byte, 0x1000)
	first := true
	createGame := &game.PowerHistoryCreateGame{}
	for {
		_, err := io.ReadFull(c.conn, lenBuf)
		if err != nil {
			panic(err)
		}
		plen := int(binary.LittleEndian.Uint32(lenBuf))
		if plen > len(packetBuf) {
			packetBuf = make([]byte, plen*3/2)
		}
		_, err = io.ReadFull(c.conn, packetBuf[:plen])
		if err != nil {
			panic(err)
		}

		log.Printf("received %d bytes from kettle: %s\n", plen, string(packetBuf[:plen]))
		packets := []KettlePacket{}
		err = json.Unmarshal(packetBuf[:plen], &packets)
		if err != nil {
			panic(err)
		}
		history := []*game.PowerHistoryData{}
		deferredOptions := []*game.AllOptions{}
		deferredEntityChoices := []*game.EntityChoices{}
		if first {
			history = append(history, &game.PowerHistoryData{
				CreateGame: createGame,
			})
			first = false
		}
		for _, packet := range packets {
			log.Printf("got a %s!\n", packet.Type)
			switch packet.Type {
			case "GameEntity":
				createGame.GameEntity = packet.GameEntity.ToProto()
			case "Player":
				player := &game.Player{}
				player.Id = proto.Int32(int32(packet.Player.EntityID - 1))
				packet.Player.Tags["53"] = packet.Player.EntityID
				packet.Player.Tags["30"] = packet.Player.EntityID - 1
				player.Entity = packet.Player.ToProto()
				createGame.Players = append(createGame.Players, player)
			case "TagChange":
				tagChange := &game.PowerHistoryTagChange{}
				tag := MakeTag(packet.TagChange.Tag, packet.TagChange.Value)
				tagChange.Tag = tag.Name
				tagChange.Value = tag.Value
				tagChange.Entity = proto.Int32(int32(packet.TagChange.EntityID))
				history = append(history, &game.PowerHistoryData{
					TagChange: tagChange,
				})
			case "ActionStart":
				start := &game.PowerHistoryStart{}
				s := packet.ActionStart
				start.Type = (*game.HistoryBlock_Type)(proto.Int32(int32(s.SubType)))
				start.Index = proto.Int32(int32(s.Index))
				start.Source = proto.Int32(int32(s.EntityID))
				start.Target = proto.Int32(int32(s.Target))
				history = append(history, &game.PowerHistoryData{
					PowerStart: start,
				})
			case "ActionEnd":
				history = append(history, &game.PowerHistoryData{
					PowerEnd: &game.PowerHistoryEnd{},
				})
			case "FullEntity":
				full := &game.PowerHistoryEntity{}
				e := packet.FullEntity.ToProto()
				full.Entity = e.Id
				full.Tags = e.Tags
				full.Name = proto.String(packet.FullEntity.CardID)
				history = append(history, &game.PowerHistoryData{
					FullEntity: full,
				})
			case "ShowEntity":
			case "HideEntity":
			case "MetaData":
			case "Choices":
			case "Options":
				options := &game.AllOptions{}
				options.Id = proto.Int32(int32(c.lastOptionID))
				c.lastOptionID++
				for _, o := range packet.Options {
					options.Options = append(options.Options, o.ToProto())
				}
				deferredOptions = append(deferredOptions, options)
			case "EntityChoices":
				deferredEntityChoices = append(deferredEntityChoices,
					packet.EntityChoices.ToProto())
			default:
				log.Panicf("unknown Kettle packet type: %s", packet.Type)
			}
		}
		c.g.OnHistory(history)
		for _, options := range deferredOptions {
			c.g.OnOptions(options)
		}
		for _, choices := range deferredEntityChoices {
			c.g.OnEntityChoices(choices)
		}
	}
}

func (c *KettleClient) writeLoop() {
	defer c.closeOnError()
	for {
		select {
		case buf := <-c.write:
			lenBuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(lenBuf, uint32(len(buf)))
			written, err := c.conn.Write(append(lenBuf, buf...))
			if err != nil {
				panic(err)
			}
			log.Printf("wrote %d bytes to kettle: %s\n", written, string(buf))
		case <-c.quit:
			return
		}
	}
}

type KettlePacket struct {
	Type           string
	GameID         string
	CreateGame     *KettleCreateGame    `json:",omitempty"`
	GameEntity     *KettleEntity        `json:",omitempty"`
	Player         *KettlePlayer        `json:",omitempty"`
	TagChange      *KettleTagChange     `json:",omitempty"`
	ActionStart    *KettleActionStart   `json:",omitempty"`
	FullEntity     *KettleFullEntity    `json:",omitempty"`
	ShowEntity     *KettleFullEntity    `json:",omitempty"`
	HideEntity     *KettleFullEntity    `json:",omitempty"`
	MetaData       *KettleMetaData      `json:",omitempty"`
	Choices        *KettleChoices       `json:",omitempty"`
	Options        []*KettleOption      `json:",omitempty"`
	SendOption     *KettleSendOption    `json:",omitempty"`
	EntityChoices  *KettleEntityChoices `json:",omitempty"`
	ChooseEntities []int                `json:",omitempty"`
	Concede        int                  `json:",omitempty"`
}

type KettleCreateGame struct {
	Players []KettleCreatePlayer
}

type KettleCreatePlayer struct {
	Name  string
	Hero  string
	Cards []string
}

type GameTags map[string]int

func TagsToProto(tags GameTags) []*game.Tag {
	res := []*game.Tag{}
	for name, value := range tags {
		nameI, err := strconv.Atoi(name)
		if err != nil {
			panic(err)
		}
		t := MakeTag(nameI, value)
		res = append(res, t)
	}
	return res
}

type KettleEntity struct {
	EntityID int
	Tags     GameTags
}

func (e *KettleEntity) ToProto() *game.Entity {
	res := &game.Entity{}
	res.Id = proto.Int32(int32(e.EntityID))
	res.Tags = TagsToProto(e.Tags)
	return res
}

func MakeTag(name, value int) *game.Tag {
	t := &game.Tag{}
	if name == 50 {
		value -= 1
	}
	t.Name = proto.Int32(int32(name))
	t.Value = proto.Int32(int32(value))
	return t
}

type KettlePlayer struct {
	KettleEntity
	PlayerID int
}

type KettleFullEntity struct {
	KettleEntity
	CardID string
}

type KettleTagChange struct {
	EntityID int
	Tag      int
	Value    int
}

type KettleActionStart struct {
	EntityID int
	SubType  int
	Index    int
	Target   int
}

type KettleMetaData struct {
	Meta int
	Data int
	Info int
}

type KettleChoices struct {
	Type     int
	EntityID int
	Source   int
	Min      int
	Max      int
	Choices  []int
}

type KettleOption struct {
	Type       int
	MainOption *KettleSubOption   `json:",omitempty"`
	SubOptions []*KettleSubOption `json:",omitempty"`
}

type KettleSubOption struct {
	ID      int
	Targets []int `json:",omitempty"`
}

type KettleSendOption struct {
	Index     int
	Target    int
	SubOption int
	Position  int
}

type KettleEntityChoices struct {
	ChoiceType int
	CountMin   int
	CountMax   int
	Entities   []int
	Source     int
	PlayerId   int
}

func (o *KettleOption) ToProto() *game.Option {
	res := &game.Option{}
	var x = game.Option_Type(o.Type)
	res.Type = &x
	if o.MainOption != nil {
		res.MainOption = o.MainOption.ToProto()
	}
	for _, s := range o.SubOptions {
		res.SubOptions = append(res.SubOptions, s.ToProto())
	}
	return res
}

func (s *KettleSubOption) ToProto() *game.SubOption {
	res := &game.SubOption{}
	res.Id = proto.Int32(int32(s.ID))
	for _, t := range s.Targets {
		res.Targets = append(res.Targets, int32(t))
	}
	return res
}

func (c *KettleEntityChoices) ToProto() *game.EntityChoices {
	res := &game.EntityChoices{}
	res.ChoiceType = proto.Int32(int32(c.ChoiceType))
	res.CountMin = proto.Int32(int32(c.CountMin))
	res.CountMax = proto.Int32(int32(c.CountMax))
	res.Source = proto.Int32(int32(c.Source))
	res.PlayerId = proto.Int32(int32(c.PlayerId))
	for _, ei := range c.Entities {
		res.Entities = append(res.Entities, int32(ei))
	}
	return res
}
