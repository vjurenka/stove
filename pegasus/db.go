package pegasus

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var db = openDB()

func openDB() gorm.DB {
	db, err := gorm.Open("sqlite3", "db/pegasus.db")
	if err != nil {
		panic(err)
	}

	db.SingularTable(true)
	return db
}

type Achieve struct {
	ID               int32
	AccountID        int64
	AchieveID        int32

	Progress         int32
	AckProgress      int32
	CompletionCount  int32
	Active           bool
	// started_count doesn't seem to be used
	DateGiven        time.Time
	DateCompleted    time.Time
	// do_not_ack is also not used
}

type Booster struct {
	ID          int64
	AccountID   int64
	BoosterType int
	Opened      bool
	Cards       []BoosterCard
}

type BoosterCard struct {
	ID        int64
	BoosterID int64
	CardID    int
	Premium   int
}

type FavoriteHero struct {
	ID           int64
	AccountID    int64
	ClassID      int32
	CardID       int32
	Premium      int32
}

type Deck struct {
	ID           int64
	AccountID    int64
	DeckType     int
	Name         string
	HeroID       int
	HeroPremium  int
	CardBackID   int32
	LastModified time.Time
	Cards        []DeckCard
}

type DeckCard struct {
	ID      int64
	DeckID  int64
	CardID  int
	Premium int
	Num     int
}

type DbfAchieve struct {
	ID          int32
	AchType     string
	Triggered   string
	AchQuota    int
	Race        int
	Reward      string
	RewardData1 int
	RewardData2 int
	CardSet     int
	Event       string
	NameEnus    string
}

type DbfCard struct {
	ID            int
	NoteMiniGuid  string
	IsCollectible bool
	NameEnus      string
	ClassID       int32
}

type DbfCardBack struct {
	ID       int32
	Data1    int
	source   string
	NameEnus string
}

type SeasonProgress struct {
	ID        int
	AccountID int64

	StarLevel            int
	Stars                int
	LevelStart, LevelEnd int
	LegendRank           int
	SeasonWins           int
	Streak               int
}

type AccountLicense struct {
	ID        int64
	AccountID int64
	LicenseID int64
}

type License struct {
	ID        int64
	ProductID int
}
