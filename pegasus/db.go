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

type Deck struct {
	ID           int64
	AccountID    int64
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
	DeckID  int64
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
