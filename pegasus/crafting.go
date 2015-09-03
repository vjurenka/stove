package pegasus

import (
	"github.com/HearthSim/hs-proto-go/pegasus/util"
	"github.com/golang/protobuf/proto"
)

type Crafting struct{}

func (s *Crafting) Init(sess *Session) {
	sess.RegisterPacket(util.BuySellCard_ID, OnCraft)
	sess.RegisterPacket(util.MassDisenchantRequest_ID, OnMassDisenchant)
}

func OnCraft(s *Session, body []byte) *Packet {
	req := util.BuySellCard{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}
	count := req.GetCount()
	// If only one card is being bought/sold, the count is 0
	if count == 0 {
		count = 1
	}
	def := req.GetDef()
	buying := req.GetBuying()

	if buying {
		card := CollectionCard{}
		if !db.Where("account_id = ? AND card_id = ? AND premium = ?", s.Account.ID, def.GetAsset(), def.GetPremium()).First(&card).RecordNotFound() {
			db.Model(&card).Update("num", card.Num+count)
		} else {
			card.AccountID = s.Account.ID
			card.CardID = def.GetAsset()
			card.Premium = def.GetPremium()
			card.Num = count
			db.Save(&card)
		}
		account := Account{}
		db.Where("id = ?", s.Account.ID).First(&account)
		db.Model(&account).Update("dust", account.Dust-int64(req.GetUnitBuyPrice()))

		res := util.BoughtSoldCard{}
		res.Def = def
		res.Amount = req.UnitBuyPrice
		result := util.BoughtSoldCard_BOUGHT
		res.Result = &result
		res.Count = proto.Int32(count)
		res.Nerfed = proto.Bool(false)
		// These values are always 0 in the response
		res.UnitBuyPrice = proto.Int32(0)
		res.UnitSellPrice = proto.Int32(0)
		return EncodePacket(util.BoughtSoldCard_ID, &res)
	} else {
		card := CollectionCard{}
		if !db.Where("account_id = ? AND card_id = ? AND premium = ?", s.Account.ID, def.GetAsset(), def.GetPremium()).First(&card).RecordNotFound() {
			db.Model(&card).Update("num", card.Num-count)
		} else {
			panic("Tried to sell a card we don't own!")
		}
		account := Account{}
		db.Where("id = ?", s.Account.ID).First(&account)
		db.Model(&account).Update("dust", account.Dust+int64(req.GetUnitBuyPrice()))

		res := util.BoughtSoldCard{}
		res.Def = def
		res.Amount = req.UnitSellPrice
		result := util.BoughtSoldCard_SOLD
		res.Result = &result
		res.Count = proto.Int32(count)
		res.Nerfed = proto.Bool(false)
		// These values are always 0 in the response
		res.UnitBuyPrice = proto.Int32(0)
		res.UnitSellPrice = proto.Int32(0)
		return EncodePacket(util.BoughtSoldCard_ID, &res)
	}
}

func OnMassDisenchant(s *Session, body []byte) *Packet {
	res := util.MassDisenchantResponse{}
	amount := int32(0)
	collection := []CollectionCard{}
	db.Where("account_id = ?", s.Account.ID).Find(&collection)
	for _, cards := range collection {
		dbfCard := DbfCard{}
		db.Where("id = ?", cards.CardID).Find(&dbfCard)
		// Legendary
		if dbfCard.Rarity == 5 && cards.Num > 1{
			if cards.Premium == 1 {
					amount += dbfCard.GoldSellPrice
					db.Model(&cards).Update("num", 1)
			} else {
				amount += dbfCard.SellPrice
				db.Model(&cards).Update("num", 1)
			}
		}
		// Other
		if dbfCard.Rarity != 5 && cards.Num > 2 {
			if cards.Premium == 1 {
					amount += dbfCard.GoldSellPrice
					db.Model(&cards).Update("num", 2)
			} else {
				amount += dbfCard.SellPrice
				db.Model(&cards).Update("num", 2)
			}
		}
	}
	res.Amount = proto.Int32(amount)
	return EncodePacket(util.MassDisenchantResponse_ID, &res)
}
