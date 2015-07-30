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
	db.AutoMigrate(
		&Account{},
		&AccountLicense{},
		&Booster{},
		&BoosterCard{},
		&Deck{},
		&DeckCard{},
		&License{},
		&SeasonProgress{})

	return db
}

type Booster struct {
	ID          int
	AccountID   int
	BoosterType int
	Opened      bool
	Cards       []BoosterCard
}

type BoosterCard struct {
	ID        int
	BoosterID int
	CardID    int
	Premium   int
}

type Deck struct {
	ID           int64
	AccountID    int
	DeckType     int
	Name         string
	HeroID       int
	HeroPremium  int
	CardBackID   int
	LastModified time.Time
	Cards        []DeckCard
}

type DeckCard struct {
	ID      int64
	DeckID  int
	CardID  int
	Premium int
}

type DbfCard struct {
	ID            int
	NoteMiniGuid  string
	IsCollectible bool
	NameEnus      string
}

type SeasonProgress struct {
	ID        int
	AccountID int

	StarLevel            int
	Stars                int
	LevelStart, LevelEnd int
	LegendRank           int
	SeasonWins           int
	Streak               int
}

type AccountLicense struct {
	ID        int
	AccountID int
	LicenseID int
}

type License struct {
	ID        int
	ProductID int
}
