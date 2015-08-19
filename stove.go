package main

import (
	"flag"
	"fmt"
	"github.com/HearthSim/stove/bnet"
	"github.com/HearthSim/stove/pegasus"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
)

const (
	CONN_DEFAULT_HOST = "localhost"
	CONN_DEFAULT_PORT = 1119
)

func migrate() {
	conn, err := gorm.Open("sqlite3", "db/pegasus.db")
	if err != nil {
		panic(err)
	}

	conn.LogMode(true)
	conn.SingularTable(true)
	conn.AutoMigrate(
		&pegasus.Account{},
		&pegasus.AccountLicense{},
		&pegasus.Achieve{},
		&pegasus.Booster{},
		&pegasus.BoosterCard{},
		&pegasus.FavoriteHero{},
		&pegasus.Deck{},
		&pegasus.DeckCard{},
		&pegasus.License{},
		&pegasus.SeasonProgress{},
		&pegasus.Bundle{},
		&pegasus.ProductGoldCost{},
		&pegasus.Product{},
	)

	conn.Close()
}

func main() {
	addr := fmt.Sprintf("%s:%d", CONN_DEFAULT_HOST, CONN_DEFAULT_PORT)
	flag.StringVar(&addr, "bind", addr, "The address to run on")
	runMigrate := flag.Bool("migrate", false, "Perform a database migration and exit")
	flag.Parse()

	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%d", addr, CONN_DEFAULT_PORT)
	}

	if *runMigrate {
		fmt.Printf("Performing database migration\n")
		migrate()
		os.Exit(0)
	}

	serv := bnet.NewServer()
	serv.RegisterGameServer("WTCG", pegasus.NewServer(serv))

	fmt.Printf("Listening on %s ...\n", addr)
	err := serv.ListenAndServe(addr)
	if err != nil {
		fmt.Println(err)
	}
}
