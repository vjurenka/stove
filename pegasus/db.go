package pegasus

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var db = openDB()

func openDB() gorm.DB {
	db, err := gorm.Open("sqlite3", "pegasus.db")
	if err != nil {
		panic(err)
	}

	db.SingularTable(true)
	db.AutoMigrate(
		&Account{},
		&AccountLicense{},
		&License{},
		&SeasonProgress{})

	return db
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
