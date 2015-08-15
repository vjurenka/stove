package pegasus

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"time"
)

func (s *Draft) Init(sess *Session) {
	sess.RegisterUtilHandler(0, 235, OnDraftBegin)
	sess.RegisterUtilHandler(0, 244, OnDraftGetPicksAndContents)
}

func MakeHeroChoices() (choices []DraftChoice) {
	favoriteHeroes := []FavoriteHero{}
	db.Limit(3).Find(&favoriteHeroes)
	for i := 1; i <= 3; i++ {
		choices = append(choices, DraftChoice{
			CardID:      favoriteHeroes[i-1].CardID,
			ChoiceIndex: i,
			Slot:        0,
		})
	}
	return choices
}

func ChoicesToCardDefs(choices []DraftChoice) (defs []*hsproto.PegasusShared_CardDef) {
	for _, choice := range choices {
		defs = append(defs, &hsproto.PegasusShared_CardDef{
			Asset:   proto.Int32(choice.CardID),
			Premium: proto.Int32(int32(0)),
		})
	}
	return defs
}

func OnDraftBegin(s *Session, body []byte) ([]byte, error) {
	deck := Deck{
		AccountID:    s.Account.ID,
		DeckType:     int(hsproto.PegasusShared_DeckType_DRAFT_DECK),
		Name:         "Arena Deck",
		CardBackID:   0, //TODO
		LastModified: time.Now().UTC(),
	}
	db.Create(&deck)

	draft := Draft{
		AccountID:   s.Account.ID,
		DeckID:      deck.ID,
		CurrentSlot: 0,
		Choices:     MakeHeroChoices(),
		Ended:       false,
	}
	db.Create(&draft)

	choiceList := ChoicesToCardDefs(draft.Choices)
	res := hsproto.PegasusUtil_DraftBeginning{
		DeckId:     proto.Int64(deck.ID),
		ChoiceList: choiceList,
	}
	return EncodeUtilResponse(246, &res)
}

func OnDraftGetPicksAndContents(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_DraftGetPicksAndContents{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}

	draft := Draft{}
	if db.Where("not ended and account_id = ?", s.Account.ID).First(&draft).RecordNotFound() {
		code := hsproto.PegasusUtil_DraftError_DE_NOT_IN_DRAFT
		res := hsproto.PegasusUtil_DraftError{
			ErrorCode: &code,
		}

		return EncodeUtilResponse(251, &res)
	}
	heroDef := MakeCardDef(834, 0)

	res := hsproto.PegasusUtil_DraftChoicesAndContents{
		DeckId:  proto.Int64(0),
		Slot:    proto.Int32(0),
		Wins:    proto.Int32(12),
		Losses:  proto.Int32(0),
		HeroDef: heroDef,
	}

	// stub
	return EncodeUtilResponse(248, &res)
}
