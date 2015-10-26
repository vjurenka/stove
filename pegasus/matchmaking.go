package pegasus

import (
	"github.com/HearthSim/hs-proto-go/bnet/entity"
	"github.com/HearthSim/hs-proto-go/bnet/game_master_types"
	"github.com/HearthSim/hs-proto-go/pegasus/shared"
	"github.com/HearthSim/stove/bnet"
	"github.com/HearthSim/stove/pegasus/game"
	"github.com/golang/protobuf/proto"
	"log"
)

func (s *Session) HandleFindGame(req map[string]interface{}) {
	gameType := shared.BnetGameType(req["type"].(int64))
	deckID := req["deck"].(int64)
	scenario := &DbfScenario{}
	db.First(scenario, int(req["scenario"].(int64)))
	if scenario.ID == 0 {
		panic("bad scenario ID")
	}
	deck := Deck{}
	db.Preload("Cards").First(&deck, deckID)
	log.Printf("handling queue for scenario %v with type %s and deck %d\n",
		*scenario, gameType.String(), deckID)
	aiDeck := Deck{}
	db.Preload("Cards").Where(&Deck{
		DeckType: int(shared.DeckType_PRECON_DECK),
		HeroID:   int32(scenario.Player2HeroCardID),
	}).First(&aiDeck)
	if scenario.Players == 1 {
		player1Cards := []string{}
		player1Premium := []bool{}
		for _, deckCard := range deck.Cards {
			for i := int32(0); i < deckCard.Num; i++ {
				player1Cards = append(player1Cards,
					cardAssetIdToMiniGuid[deckCard.CardID])
				player1Premium = append(player1Premium, false)
			}
		}
		player2Cards := []string{}
		player2Premium := []bool{}
		for _, deckCard := range aiDeck.Cards {
			for i := int32(0); i < deckCard.Num; i++ {
				player2Cards = append(player2Cards,
					cardAssetIdToMiniGuid[deckCard.CardID])
				player2Premium = append(player2Premium, false)
			}
		}
		params := &game.GameStartInfo{}
		params.Players = append(params.Players, game.PlayerInfo{
			DisplayName: s.Account.displayName,
			GameAccountId: &shared.BnetId{
				Hi: proto.Uint64(GameAccountEntityIDHi),
				Lo: proto.Uint64(uint64(s.Account.ID)),
			},
			HeroCardId: cardAssetIdToMiniGuid[deck.HeroID],
			CardIds:    player1Cards,
			Premium:    player1Premium,
		})
		params.Players = append(params.Players, game.PlayerInfo{
			DisplayName: "The Innkeeper",
			GameAccountId: &shared.BnetId{
				Hi: proto.Uint64(GameAccountEntityIDHi),
				Lo: proto.Uint64(0),
			},
			HeroCardId: cardAssetIdToMiniGuid[aiDeck.HeroID],
			CardIds:    player2Cards,
			Premium:    player2Premium,
		})
		g := game.CreateGame(params)
		connectInfo := &game_master_types.ConnectInfo{}
		// TODO: figure out the right host
		connectInfo.Host = proto.String("127.0.0.1")
		connectInfo.Port = proto.Int32(int32(g.Address.Port))
		connectInfo.Token = []byte(g.Players[0].Password)
		connectInfo.MemberId = &entity.EntityId{}
		connectInfo.MemberId.High = proto.Uint64(0)
		connectInfo.MemberId.Low = proto.Uint64(0)

		connectInfo.Attribute = bnet.NewNotification("", map[string]interface{}{
			"id":                 g.Players[0].ClientHandle,
			"game":               g.GameHandle,
			"resumable":          false,
			"spectator_password": g.SpectatorPassword,
			"version":            "3.0.0.10604",
		}).Attributes
		buf, err := proto.Marshal(connectInfo)
		if err != nil {
			panic(err)
		}
		defer func() {
			s.gameNotifications <- bnet.NewNotification(
				bnet.NotifyQueueResult,
				map[string]interface{}{
					"connection_info": bnet.MessageValue{buf},
					"targetId":        *bnet.EntityId(0, 0),
					"forwardToClient": true,
				})
		}()
	} else {
		panic(nyi)
	}
	s.gameNotifications <- bnet.NewNotification(bnet.NotifyFindGameResponse,
		map[string]interface{}{
			"queued":    true,
			"requestId": uint(1),
		})
}
