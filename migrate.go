package main

import (
	"github.com/HearthSim/stove/pegasus"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	conn, err := gorm.Open("sqlite3", "db/pegasus.db")
	if err != nil {
		panic(err)
	}

	conn.LogMode(true)
	conn.SingularTable(true)
	conn.AutoMigrate(
		&pegasus.Account{},
		&pegasus.AccountLicense{},
		&pegasus.Booster{},
		&pegasus.BoosterCard{},
		&pegasus.Deck{},
		&pegasus.DeckCard{},
		&pegasus.License{},
		&pegasus.SeasonProgress{})

	conn.Close()
}
