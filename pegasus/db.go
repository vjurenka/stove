package pegasus

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

var db = openDB()

func openDB() gorm.DB {
	dbFile := os.Getenv("PEGASUS_DB")
	if len(dbFile) == 0 {
		dbFile = "db/pegasus.db"
	}
	db, err := gorm.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}

	db.SingularTable(true)
	return db
}

func Migrate() {
	db.LogMode(true)
	err := db.AutoMigrate(
		&Account{},
		&AccountLicense{},
		&Achieve{},
		&Booster{},
		&BoosterCard{},
		&FavoriteHero{},
		&Deck{},
		&DeckCard{},
		&License{},
		&SeasonProgress{},
		&Bundle{},
		&ProductGoldCost{},
		&Product{},
		&Draft{},
		&DraftChoice{},
		&CollectionCard{},
	).Error

	if err != nil {
		panic(err)
	}
}

type Account struct {
	ID        int64
	BnetID    int
	Gold      int64
	Dust      int64
	UpdatedAt time.Time
	Flags     int64

	Progress []SeasonProgress
	Licenses []License
}

type Achieve struct {
	ID        int32
	AccountID int64
	AchieveID int32

	Progress        int32
	AckProgress     int32
	CompletionCount int32
	Active          bool
	// started_count doesn't seem to be used
	DateGiven     time.Time
	DateCompleted time.Time
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
	CardID    int32
	Premium   int32
}

type FavoriteHero struct {
	ID        int64
	AccountID int64
	ClassID   int32
	CardID    int32
	Premium   int32
}

type Deck struct {
	ID           int64
	AccountID    int64
	DeckType     int
	Name         string
	HeroID       int32
	HeroPremium  int32
	CardBackID   int32
	LastModified time.Time
	Cards        []DeckCard
}

type DeckCard struct {
	ID      int64
	DeckID  int64
	CardID  int32
	Premium int32
	Num     int32
}

type Draft struct {
	ID          int64
	AccountID   int64
	DeckID      int64
	Wins        int32
	Losses      int32
	CreatedAt   time.Time
	EndedAt     time.Time
	Ended       bool
	Choices     []DraftChoice
	CurrentSlot int32
	// PurchasedWithGold bool
}

type DraftChoice struct {
	ID          int
	DraftID     int64
	CardID      int32
	ChoiceIndex int
	Slot        int32
	//Premium int32
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
	ID            int32
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

type Bundle struct {
	ID        int64
	ProductID string
	AppleID   string
	AmazonID  string
	GoogleID  string
	Items     []Product `gorm:"many2many:bundle_products;"`
	EventName string
}

type ProductGoldCost struct {
	ID          int64
	ProductType int
	PackType    int32
	Cost        int64
}

type Product struct {
	ID          int64
	ProductType int
	ProductData int32
	Quantity    int32
}

type CollectionCard struct {
	ID        int64
	AccountID int64
	CardID    int32
	Premium   int32
	Num       int32
}
