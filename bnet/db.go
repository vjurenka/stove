package bnet

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

var db = openDB()

func openDB() gorm.DB {
	dbFile := os.Getenv("BNET_DB")
	if len(dbFile) == 0 {
		dbFile = "db/bnet.db"
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
		&Friend{},
		&InvitationRequest{},
	).Error

	if err != nil {
		panic(err)
	}
}

// EntityIDs are a 128-bit GUID applied to various entities in the bnet system.
// The high 64 bits of EntityID are generally fixed per the application.

// Byte 15 of an EntityID signifies which application uses it, bnet or a game.
const (
	EntityIDKindNone = iota
	EntityIDKindAccount
	EntityIDKindGameAccount
	EntityIDKindChannel = 0x6
)

// Byte 12 is used for the region; bnet accounts are not region-specific, but
// game accounts may be.
const (
	EntityIDRegionNone = iota
	EntityIDRegionNorthAmerica
	EntityIDRegionEurope
	// Use this special region to distinguish stove's game account entity ids
	// from entity ids on other systems.
	EntityIDRegionTest = 0x7A
)

// Bytes 11 through 8 are used for the FourCC of the game in question.
const (
	EntityIDGameNone    = 0
	EntityIDGameApp     = 0x417070   // 'App', the bnet app
	EntityIDGamePegasus = 0x57544347 // 'WTCG', i.e. pegasus
)

// This is the high part of all account entity ids in this server
const BnetAccountEntityIDHi uint64 = (EntityIDKindAccount << 56) |
	(EntityIDRegionTest << 32) |
	(EntityIDGameNone)
const BnetGameAccountEntityIDHi uint64 = (EntityIDKindGameAccount << 56) |
	(EntityIDRegionTest << 32) |
	(EntityIDGamePegasus)

type Account struct {
	// The lo part of the full entity id
	ID            uint64
	Email         string
	WebCredential string
	// Formatted as Name#1234
	BattleTag    string
	Flags        int64
	GameAccounts []GameAccount
}

type GameAccount struct {
	// This ID must be the same as the account ID in the game's db.
	ID        int64
	CreatedAt time.Time
	Game      string
	Region    string
}

type AccountGameAccount struct {
	ID            int64
	AccountID     int64
	GameAccountID int64
}

type Friend struct {
	ID     int64
	Source uint64 // pointing to Account table
	Target uint64 // -""-
}

type InvitationRequest struct {
	ID             uint64 // this is requestID, it is unique ID that represents friend request
	InviterID      uint64
	InviteeID      uint64
	CreationTime   time.Time
	ExpirationTime time.Time
}
