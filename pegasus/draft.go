package pegasus

import (
	"github.com/HearthSim/hs-proto-go/pegasus/shared"
	"github.com/HearthSim/hs-proto-go/pegasus/util"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

func (s *Draft) Init(sess *Session) {
	sess.RegisterPacket(util.DraftBegin_ID, OnDraftBegin)
	sess.RegisterPacket(util.DraftGetPicksAndContents_ID, OnDraftGetPicksAndContents)
	sess.RegisterPacket(util.DraftRetire_ID, OnDraftRetire)
	sess.RegisterPacket(util.DraftMakePick_ID, OnDraftMakePick)
	sess.RegisterPacket(util.DraftAckRewards_ID, OnDraftAckRewards)
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

func MakeCardChoices(slot int32) (choices []DraftChoice) {
	cards := []DbfCard{}
	// just use first 3 classic set cards as a placeholder
	db.Limit(3).Where("is_collectible = ? and note_mini_guid GLOB ?", 1, "CS[12]_[0-9][0-9][0-9]").Find(&cards)
	for i := 1; i <= 3; i++ {
		choices = append(choices, DraftChoice{
			CardID:      cards[i-1].ID,
			ChoiceIndex: i,
			Slot:        slot,
		})
	}
	return choices
}

func ChoicesToCardDefs(choices []DraftChoice) (defs []*shared.CardDef) {
	for _, choice := range choices {
		defs = append(defs, &shared.CardDef{
			Asset:   proto.Int32(choice.CardID),
			Premium: proto.Int32(int32(0)),
		})
	}
	return defs
}

func OnDraftBegin(s *Session, body []byte) *Packet {
	deck := Deck{
		AccountID:    s.Account.ID,
		DeckType:     int(shared.DeckType_DRAFT_DECK),
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
	res := util.DraftBeginning{
		DeckId:     proto.Int64(deck.ID),
		ChoiceList: choiceList,
	}
	return EncodePacket(util.DraftBeginning_ID, &res)
}

func OnDraftGetPicksAndContents(s *Session, body []byte) *Packet {
	req := util.DraftGetPicksAndContents{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}

	draft := Draft{}
	if db.Where("not ended and account_id = ?", s.Account.ID).First(&draft).RecordNotFound() {
		code := util.DraftError_DE_NOT_IN_DRAFT
		res := util.DraftError{
			ErrorCode: &code,
		}

		return EncodePacket(util.DraftError_ID, &res)
	}

	choices := []DraftChoice{}
	db.Where("draft_id = ?", draft.ID).Find(&choices)
	choiceList := ChoicesToCardDefs(choices)

	deck := Deck{}
	db.Where("id = ?", draft.DeckID).Preload("Cards").First(&deck)
	heroDef := shared.CardDef{
		Asset:   proto.Int32(deck.HeroID),
		Premium: proto.Int32(0),
	}
	cards := []*shared.DeckCardData{}
	for i, card := range deck.Cards {
		cards = append(cards, &shared.DeckCardData{
			Def:    MakeCardDef(card.CardID, 0),
			Handle: proto.Int32(int32(i)),
			Qty:    proto.Int32(0),
			Prev:   proto.Int32(int32(i) - 1),
		})
	}

	res := util.DraftChoicesAndContents{
		DeckId:     proto.Int64(draft.DeckID),
		Slot:       proto.Int32(draft.CurrentSlot),
		Wins:       proto.Int32(draft.Wins),
		Losses:     proto.Int32(draft.Losses),
		Cards:      cards,
		ChoiceList: choiceList,
		HeroDef:    &heroDef,
	}

	return EncodePacket(util.DraftChoicesAndContents_ID, &res)
}

func OnDraftMakePick(s *Session, body []byte) *Packet {
	req := util.DraftMakePick{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}

	draft := Draft{}
	if db.Where("not ended and account_id = ?", s.Account.ID).First(&draft).RecordNotFound() {
		log.Panicf("received OnDraftMakePick for account with no active draft")
	}
	if req.GetSlot() != draft.CurrentSlot {
		log.Panicf("received OnDraftMakePick for the wrong slot")
	}
	if req.GetDeckId() != draft.DeckID {
		log.Panicf("received OnDraftMakePick for the wrong deck")
	}
	pick := DraftChoice{}
	db.Where("draft_id = ? and choice_index = ?", draft.ID, req.GetIndex()).First(&pick)
	db.Where("draft_id = ?", draft.ID).Delete(&DraftChoice{})

	if draft.CurrentSlot == 0 {
		deck := Deck{}
		db.Where("id = ?", draft.DeckID).First(&deck)
		deck.HeroID = pick.CardID
		deck.HeroPremium = 0
		deck.LastModified = time.Now().UTC()
		db.Save(&deck)
	} else {
		card := DeckCard{
			DeckID:  draft.DeckID,
			CardID:  pick.CardID,
			Premium: 0,
			Num:     1,
		}
		db.Save(&card)
	}

	if draft.CurrentSlot < 30 {
		draft.Choices = MakeCardChoices(draft.CurrentSlot)
	}
	draft.CurrentSlot += 1
	db.Save(&draft)

	choices := []*shared.CardDef{}
	for _, choice := range draft.Choices {
		choices = append(choices, &shared.CardDef{
			Asset:   proto.Int32(choice.CardID),
			Premium: proto.Int32(int32(0)),
		})
	}
	res := util.DraftChosen{
		Chosen:         MakeCardDef(pick.CardID, 0),
		NextChoiceList: choices,
	}

	return EncodePacket(util.DraftChosen_ID, &res)
}

func MakeChest() (chest shared.RewardChest) {
	// TODO take arguments to determine the contents
	// There are up to 5 bags each which can hold a booster, card, dust or gold

	chest.Bag1 = &shared.RewardBag{
		RewardBooster: &shared.ProfileNoticeRewardBooster{
			BoosterType:  proto.Int32(1),
			BoosterCount: proto.Int32(1),
		},
	}
	chest.Bag2 = &shared.RewardBag{
		RewardCard: &shared.ProfileNoticeRewardCard{
			Card:     MakeCardDef(2078, 1),
			Quantity: proto.Int32(1),
		},
	}
	chest.Bag3 = &shared.RewardBag{
		RewardDust: &shared.ProfileNoticeRewardDust{
			Amount: proto.Int32(69),
		},
	}
	chest.Bag4 = &shared.RewardBag{
		RewardGold: &shared.ProfileNoticeRewardGold{
			Amount: proto.Int32(42),
		},
	}
	return chest
}

func OnDraftRetire(s *Session, body []byte) *Packet {
	req := util.DraftRetire{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}

	draft := Draft{}
	if db.Where("not ended and account_id = ?", s.Account.ID).First(&draft).RecordNotFound() {
		log.Panicf("received OnDraftRetire for account with no active draft")
	}
	draft.Ended = true
	draft.EndedAt = time.Now().UTC()
	db.Save(&draft)

	chest := MakeChest()
	res := util.DraftRetired{
		DeckId: req.DeckId,
		Chest:  &chest,
	}
	return EncodePacket(util.DraftRetired_ID, &res)
}

func OnDraftAckRewards(s *Session, body []byte) *Packet {
	req := util.DraftAckRewards{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}

	res := util.DraftRewardsAcked{
		DeckId: req.DeckId,
	}
	return EncodePacket(util.DraftRewardsAcked_ID, &res)
}
