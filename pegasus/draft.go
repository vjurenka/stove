package pegasus

import (
	"fmt"
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
)

func (s *Draft) Init(sess *Session) {
	sess.RegisterUtilHandler(0, 244, OnDraftGetPicksAndContents)
}

func GetHeroChoices(choices []*hsproto.PegasusShared_CardDef) {
	for i := 0; i < 3; i++ {
		cardId := fmt.Sprintf("HERO_0%d", i)
		card := DbfCard{}
		db.Where("card_id = ?", cardId).First(&card)
		choices = append(choices, MakeCardDef(card.ID, 0))
	}
}

func OnDraftGetPicksAndContents(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_DraftGetPicksAndContents{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}

	draft := Draft{}
	deck := Deck{}
	choices := []*hsproto.PegasusShared_CardDef{}
	var heroDef hsproto.PegasusShared_CardDef
	var hasCurrentDraft int32

	db.Model(&draft).Where("ended is null").Count(&hasCurrentDraft)
	if hasCurrentDraft == 0 {
		deck.DeckType = int(hsproto.PegasusShared_DeckType_DRAFT_DECK)
		deck.AccountID = 1
		db.Create(&deck)
		draft.AccountID = 1
		draft.DeckID = deck.ID
		heroDef.Asset = proto.Int32(0)
		heroDef.Premium = proto.Int32(0)
	} else {
		db.Where("ended is null").First(&draft)
		db.Where("deck_id = ?", draft.DeckID).First(&deck)
		heroDef.Asset = proto.Int32(int32(deck.HeroID))
		heroDef.Premium = proto.Int32(int32(deck.HeroPremium))
		GetHeroChoices(choices)
	}

	res := hsproto.PegasusUtil_DraftChoicesAndContents{
		DeckId:     proto.Int64(draft.DeckID),
		Slot:       proto.Int32(0),
		Wins:       proto.Int32(draft.Wins),
		Losses:     proto.Int32(draft.Losses),
		ChoiceList: choices,
		HeroDef:    &heroDef,
	}

	// stub
	return EncodeUtilResponse(248, &res)
}
