package pegasus

import (
	"github.com/HearthSim/hs-proto-go/pegasus/util"
	"github.com/golang/protobuf/proto"
)

type Store struct{}

func (s *Store) Init(sess *Session) {
	sess.RegisterPacket(util.GetBattlePayConfig_ID, OnGetBattlePayConfig)
	sess.RegisterPacket(util.GetBattlePayStatus_ID, OnGetBattlePayStatus)
	sess.RegisterPacket(util.PurchaseWithGold_ID, OnPurchaseWithGold)
}

func OnGetBattlePayConfig(s *Session, body []byte) *Packet {
	res := util.BattlePayConfigResponse{}
	// Hardcode US Dollars until we setup the DB to handle other currencies
	res.Currency = proto.Int32(1)
	res.Unavailable = proto.Bool(false)
	res.SecsBeforeAutoCancel = proto.Int32(10)

	product := ProductGoldCost{}
	db.Where("product_type = ?", 2).Find(&product)
	res.GoldCostArena = proto.Int64(product.Cost)

	goldCostBoosters := []*util.GoldCostBooster{}
	cost := []ProductGoldCost{}
	db.Where("pack_type != ?", 0).Find(&cost)
	for _, costs := range cost {
		goldCostBoosters = append(goldCostBoosters, &util.GoldCostBooster{
			Cost:     proto.Int64(costs.Cost),
			PackType: proto.Int32(costs.PackType),
		})
	}
	res.GoldCostBoosters = goldCostBoosters

	bundles := []Bundle{}
	db.Find(&bundles)
	for _, bundle := range bundles {
		bundleItems := []*util.BundleItem{}
		products := []Product{}
		db.Model(&bundle).Association("Items").Find(&products)
		for _, items := range products {
			productType := util.ProductType(items.ProductType)
			bundleItems = append(bundleItems, &util.BundleItem{
				ProductType: &productType,
				Data:        proto.Int32(items.ProductData),
				Quantity:    proto.Int32(items.Quantity),
			})
		}
		res.Bundles = append(res.Bundles, &util.Bundle{
			Id: proto.String(bundle.ProductID),
			// Hardcode $1 until price data is implemented in DB
			Cost:         proto.Float64(1.00),
			AppleId:      proto.String(bundle.AppleID),
			AmazonId:     proto.String(bundle.AmazonID),
			GooglePlayId: proto.String(bundle.GoogleID),
			// Hardcode 100 until price data is implemented in DB
			GoldCost:         proto.Int64(100),
			ProductEventName: proto.String(bundle.EventName),
			Items:            bundleItems,
		})
	}
	return EncodePacket(util.BattlePayConfigResponse_ID, &res)
}

func OnGetBattlePayStatus(s *Session, body []byte) *Packet {
	res := util.BattlePayStatusResponse{}
	status := util.BattlePayStatusResponse_PS_READY
	res.Status = &status
	res.BattlePayAvailable = proto.Bool(true)
	return EncodePacket(util.BattlePayStatusResponse_ID, &res)
}

func OnPurchaseWithGold(s *Session, body []byte) *Packet {
	req := util.PurchaseWithGold{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}

	product := ProductGoldCost{}
	productType := req.GetProduct()
	data := req.GetData()
	// If data is > 0, we're buying a pack
	if data > 0 {
		db.Where("product_type = ? AND pack_type = ?", productType, data).Find(&product)
	} else {
		db.Where("product_type = ?", productType).Find(&product)
	}

	res := util.PurchaseWithGoldResponse{}
	// TODO: Query the DB to ensure we have enough gold
	result := util.PurchaseWithGoldResponse_PR_SUCCESS
	res.Result = &result
	res.GoldUsed = proto.Int64(product.Cost * int64(req.GetQuantity()))
	return EncodePacket(util.PurchaseWithGoldResponse_ID, &res)
}
