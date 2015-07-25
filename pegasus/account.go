package pegasus

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

type Account struct {
	ID        int
	BnetID    int
	UpdatedAt time.Time
	Flags     int64

	Progress []SeasonProgress
	Licenses []License
}

func (v *Account) Init(sess *Session) {
	sess.RegisterUtilHandler(0, 201, OnGetAccountInfo)
	sess.RegisterUtilHandler(0, 205, OnUpdateLogin)
	sess.RegisterUtilHandler(0, 223, OnAckCardSeen)
	sess.RegisterUtilHandler(0, 239, OnSetOptions)
	sess.RegisterUtilHandler(0, 240, OnGetOptions)
	sess.RegisterUtilHandler(0, 253, OnGetAchieves)
	sess.RegisterUtilHandler(0, 267, OnCheckAccountLicenses)
	sess.RegisterUtilHandler(1, 276, OnCheckGameLicenses)
	sess.RegisterUtilHandler(0, 284, OnValidateAchieve)
	sess.RegisterUtilHandler(0, 305, OnGetAdventureProgress)
}

func OnAckCardSeen(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_UpdateLogin{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("FIXME: AckCardSeen = %s", req.String())
	return nil, nil
}

func OnCheckAccountLicenses(s *Session, body []byte) ([]byte, error) {
	return OnCheckLicenses(true)
}

func OnCheckGameLicenses(s *Session, body []byte) ([]byte, error) {
	return OnCheckLicenses(false)
}

func OnCheckLicenses(accountLevel bool) ([]byte, error) {
	res := hsproto.PegasusUtil_CheckLicensesResponse{}
	res.AccountLevel = proto.Bool(accountLevel)
	res.Success = proto.Bool(true)
	return EncodeUtilResponse(277, &res)
}

func OnUpdateLogin(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_UpdateLogin{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	res := hsproto.PegasusUtil_UpdateLoginComplete{}
	return EncodeUtilResponse(307, &res)
}

func OnGetAccountInfo(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_GetAccountInfo{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	switch req.Request.String() {
	case "CAMPAIGN_INFO":
		res := hsproto.PegasusUtil_ProfileProgress{}
		res.Progress = proto.Int64(6)  // ILLIDAN_COMPLETE
		res.BestForge = proto.Int32(0) // Arena wins
		return EncodeUtilResponse(233, &res)
	case "BOOSTERS":
		res := hsproto.PegasusUtil_BoosterList{}
		return EncodeUtilResponse(224, &res)
	case "FEATURES":
		res := hsproto.PegasusUtil_GuardianVars{}
		res.ShowUserUI = proto.Int32(1)
		return EncodeUtilResponse(264, &res)
	case "MEDAL_INFO":
		res := hsproto.PegasusUtil_MedalInfo{}
		res.SeasonWins = proto.Int32(0)
		res.Stars = proto.Int32(2)
		res.Streak = proto.Int32(0)
		res.StarLevel = proto.Int32(1)
		res.LevelStart = proto.Int32(1)
		res.LevelEnd = proto.Int32(3)
		res.CanLose = proto.Bool(false)
		return EncodeUtilResponse(232, &res)
	case "MEDAL_HISTORY":
		res := hsproto.PegasusUtil_MedalHistory{}
		for i := int32(1); i <= 3; i++ {
			info := &hsproto.PegasusUtil_MedalHistoryInfo{}
			info.When = PegasusDate(time.Date(2015, 8, 1, 7, 0, 0, 0, time.UTC))
			info.Season = proto.Int32(i)
			info.Stars = proto.Int32(0)
			info.StarLevel = proto.Int32(0)
			info.LevelStart = proto.Int32(0)
			info.LevelEnd = proto.Int32(0)
			info.LegendRank = proto.Int32(1)
			res.Medals = append(res.Medals, info)
		}
		return EncodeUtilResponse(234, &res)
	case "NOTICES":
		res := hsproto.PegasusUtil_ProfileNotices{}
		return EncodeUtilResponse(212, &res)
	case "DECK_LIST":
		res := hsproto.PegasusUtil_DeckList{}
		for i := 2; i <= 10; i++ {
			info := &hsproto.PegasusShared_DeckInfo{}
			info.Id = proto.Int64(int64(1000 + i))
			info.Name = proto.String("precon")
			info.CardBack = proto.Int32(0)
			info.Hero = proto.Int32(int32(heroIdToAssetId[i]))
			precon := hsproto.PegasusShared_DeckType_PRECON_DECK
			info.DeckType = &precon
			info.Validity = proto.Uint64(31)
			info.HeroPremium = proto.Int32(0)
			info.CardBackOverride = proto.Bool(false)
			info.HeroOverride = proto.Bool(false)
			res.Decks = append(res.Decks, info)
		}
		return EncodeUtilResponse(202, &res)
	case "COLLECTION":
		res := hsproto.PegasusUtil_Collection{}
		dbfCards := []DbfCard{}
		db.Where("is_collectible").Find(&dbfCards)
		for _, card := range dbfCards {
			stack1 := &hsproto.PegasusShared_CardStack{}
			stack1.LatestInsertDate = PegasusDate(time.Now().UTC())
			stack1.NumSeen = proto.Int32(2)
			stack1.Count = proto.Int32(2)
			carddef := &hsproto.PegasusShared_CardDef{}
			carddef.Asset = proto.Int32(int32(card.ID))
			carddef.Premium = proto.Int32(0)
			stack1.CardDef = carddef
			res.Stacks = append(res.Stacks, stack1)
		}
		return EncodeUtilResponse(207, &res)
	case "DECK_LIMIT":
		res := hsproto.PegasusUtil_ProfileDeckLimit{}
		res.DeckLimit = proto.Int32(9)
		return EncodeUtilResponse(231, &res)
	case "CARD_VALUES":
		res := hsproto.PegasusUtil_CardValues{}
		res.CardNerfIndex = proto.Int32(0)
		return EncodeUtilResponse(260, &res)
	case "ARCANE_DUST_BALANCE":
		res := hsproto.PegasusUtil_ArcaneDustBalance{}
		res.Balance = proto.Int64(10000)
		return EncodeUtilResponse(262, &res)
	case "GOLD_BALANCE":
		res := hsproto.PegasusUtil_GoldBalance{}
		res.Cap = proto.Int64(999999)
		res.CapWarning = proto.Int64(2000)
		res.CappedBalance = proto.Int64(1234)
		res.BonusBalance = proto.Int64(0)
		return EncodeUtilResponse(278, &res)
	case "HERO_XP":
		res := hsproto.PegasusUtil_HeroXP{}
		for i := 2; i <= 10; i++ {
			info := &hsproto.PegasusUtil_HeroXPInfo{}
			level := 2*i + 5
			maxXp := 60 + level*10
			info.ClassId = proto.Int32(int32(i))
			info.Level = proto.Int32(int32(level))
			info.CurrXp = proto.Int64(int64(maxXp / 2))
			info.MaxXp = proto.Int64(int64(maxXp))
			res.XpInfos = append(res.XpInfos, info)
		}
		return EncodeUtilResponse(283, &res)
	case "NOT_SO_MASSIVE_LOGIN":
		res := hsproto.PegasusUtil_NotSoMassiveLoginReply{}
		return EncodeUtilResponse(300, &res)
	case "REWARD_PROGRESS":
		res := hsproto.PegasusUtil_RewardProgress{}
		nextMonth := time.Date(2015, 8, 1, 7, 0, 0, 0, time.UTC)
		res.SeasonEnd = PegasusDate(nextMonth)
		res.WinsPerGold = proto.Int32(3)
		res.GoldPerReward = proto.Int32(10)
		res.MaxGoldPerDay = proto.Int32(100)
		res.SeasonNumber = proto.Int32(21)
		res.XpSoloLimit = proto.Int32(60)
		res.MaxHeroLevel = proto.Int32(60)
		res.NextQuestCancel = PegasusDate(time.Now().UTC())
		res.EventTimingMod = proto.Float32(0.291667)
		return EncodeUtilResponse(271, &res)
	case "PVP_QUEUE":
		res := hsproto.PegasusUtil_PlayQueue{}
		queue := hsproto.PegasusShared_PlayQueueInfo{}
		gametype := hsproto.PegasusShared_BnetGameType_BGT_NORMAL
		queue.GameType = &gametype
		res.Queue = &queue
		return EncodeUtilResponse(286, &res)

	case "PLAYER_RECORD":
		res := hsproto.PegasusUtil_PlayerRecords{}
		return EncodeUtilResponse(270, &res)
	case "CARD_BACKS":
		res := hsproto.PegasusUtil_CardBacks{}
		res.DefaultCardBack = proto.Int32(13)
		res.CardBacks = []int32{0, 13, 24}
		return EncodeUtilResponse(236, &res)
	case "FAVORITE_HEROES":
		res := hsproto.PegasusUtil_FavoriteHeroesResponse{}
		for i := 2; i <= 10; i++ {
			fav := &hsproto.PegasusShared_FavoriteHero{}
			fav.ClassId = proto.Int32(int32(i))
			carddef := &hsproto.PegasusShared_CardDef{}
			carddef.Asset = proto.Int32(int32(heroIdToAssetId[i]))
			fav.Hero = carddef
			res.FavoriteHeroes = append(res.FavoriteHeroes, fav)
		}
		return EncodeUtilResponse(318, &res)
	case "ACCOUNT_LICENSES":
		res := hsproto.PegasusUtil_AccountLicensesInfoResponse{}
		return EncodeUtilResponse(325, &res)
	case "BOOSTER_TALLY":
		res := hsproto.PegasusUtil_BoosterTallyList{}
		return EncodeUtilResponse(313, &res)
	default:

		return nil, nyi
	}
}

func OnGetAdventureProgress(s *Session, body []byte) ([]byte, error) {
	res := hsproto.PegasusUtil_AdventureProgressResponse{}
	return EncodeUtilResponse(306, &res)
}

func OnSetOptions(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_SetOptions{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	return nil, nil
}

func OnGetOptions(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_GetOptions{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	res := hsproto.PegasusUtil_ClientOptions{}
	res.Options = append(res.Options, &hsproto.PegasusUtil_ClientOption{
		Index:    proto.Int32(1),
		AsUint64: proto.Uint64(0x20FFFF3FFFCCFCFF),
	})
	res.Options = append(res.Options, &hsproto.PegasusUtil_ClientOption{
		Index:    proto.Int32(2),
		AsUint64: proto.Uint64(0xF0BFFFEF3FFF),
	})
	res.Options = append(res.Options, &hsproto.PegasusUtil_ClientOption{
		Index:   proto.Int32(18),
		AsInt64: proto.Int64(0xB765A8C),
	})
	return EncodeUtilResponse(241, &res)
}

func OnGetAchieves(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_GetAchieves{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	res := hsproto.PegasusUtil_Achieves{}
	for i := 1; i <= 9; i++ {
		info := &hsproto.PegasusUtil_Achieve{}
		info.Id = proto.Int32(int32(i))
		info.Progress = proto.Int32(1)
		info.AckProgress = proto.Int32(1)
		info.CompletionCount = proto.Int32(1)
		info.StartedCount = proto.Int32(1)
		info.DateGiven = PegasusDate(time.Now())
		res.List = append(res.List, info)
	}
	return EncodeUtilResponse(252, &res)
}

func OnValidateAchieve(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_ValidateAchieve{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	res := hsproto.PegasusUtil_ValidateAchieveResponse{}
	res.Achieve = proto.Int32(req.GetAchieve())
	return EncodeUtilResponse(285, &res)
}

func PegasusDate(t time.Time) *hsproto.PegasusShared_Date {
	return &hsproto.PegasusShared_Date{
		Year:  proto.Int32(int32(t.Year())),
		Month: proto.Int32(int32(t.Month())),
		Day:   proto.Int32(int32(t.Day())),
		Hours: proto.Int32(int32(t.Hour())),
		Min:   proto.Int32(int32(t.Minute())),
		Sec:   proto.Int32(int32(t.Second())),
	}
}

// A map from TAG_CLASS ids to DBF ids
var heroIdToAssetId = map[int]int{
	2:  274,
	3:  31,
	4:  637,
	5:  671,
	6:  813,
	7:  930,
	8:  1066,
	9:  893,
	10: 7,
}
