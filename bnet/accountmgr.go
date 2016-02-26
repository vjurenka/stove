package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/entity"
	"github.com/HearthSim/hs-proto-go/bnet/attribute"
	_ "time"
	"log"
	_ "fmt"
)

type amEntityId struct {
	High	uint64
	Low		uint64
}

func (e *amEntityId) GetHigh() uint64 {
	if e != nil {
		return e.High
	}
	return 0
}
func (e *amEntityId) GetLow() uint64 {
	if e != nil {
		return e.Low
	}
	return 0
}

func (e *amEntityId) String() string { return "x"+string(e.High)+","+string(e.Low)+"x" }

func (e *amEntityId) FromEntityId(source *entity.EntityId) *amEntityId {
	if source != nil {
		return &amEntityId{source.GetHigh(), source.GetLow()}
	}
	return nil
}

type amPresenceKey struct {
	Program uint32
	Group uint32
	Field uint32
	Index uint64
}

type amAccount struct {
	EntityId amEntityId
	BattleTag string
	Session *Session
	Subscribers []*amAccount
	PresenceData map[amPresenceKey]attribute.Variant
}

type amGameAccount struct {
	EntityId amEntityId
	Subscribers []*amAccount
}

type AccountManager struct {
	Accounts map[amEntityId]*amAccount
	GameAccounts map[amEntityId]*amGameAccount
	BattleTags map[string]amEntityId
}

func NewAccountManager() *AccountManager {
	log.Print("AccountManager: initializing")
	return &AccountManager{
		Accounts: map[amEntityId]*amAccount{},
		GameAccounts: map[amEntityId]*amGameAccount{},
		BattleTags: map[string]amEntityId{},
	}
}

func (am *AccountManager) GetBattleTag(e amEntityId) string {
	v, ok := am.Accounts[e]
	if !ok {
		return "Unknown"
	}
	return v.BattleTag
}

func (am *AccountManager) AddAccount(high uint64, low uint64, battletag string, session *Session) bool {
	_, ok := am.Accounts[amEntityId{high, low}]
	if !ok {
		// check low == 1 (Player account)
		if low != 1 {
			log.Printf("AccountManager: Cannot add non-account class entry")
			return false
		}
		// account doesn't exist yet
		am.Accounts[amEntityId{high, low}] = &amAccount{amEntityId{high, low}, battletag, session, []*amAccount{}, map[amPresenceKey]attribute.Variant{}}
		am.BattleTags[battletag] = amEntityId{high, low} 
		return true
	}
	return false
}

func (am *AccountManager) AddGameAccount(high uint64, low uint64) bool {
	_, ok := am.GameAccounts[amEntityId{high, low}]
	if !ok {
		// check low == 2 (Game account)
		if low != 2 {
			log.Printf("AccountManager: Cannot add non-game class entry")
			return false
		}
		// account doesn't exist yet
		am.GameAccounts[amEntityId{high, low}] = &amGameAccount{amEntityId{high, low}, []*amAccount{}}
		return true
	}
	return false
}

func (am *AccountManager) UpdatePresenceData(k amPresenceKey, v attribute.Variant) bool {
	log.Printf("AccountManager: UpdatePresenceData: %s = %s", k, v)

	return True
}
func (am *AccountManager) RemovePresenceData(k amPresenceKey) bool {
	log.Printf("AccountManager: RemovePresenceData: %s", k)
	return True
}

func (am AccountManager) Dump() {
	if len(am.Accounts) == 0{
		log.Printf("AccountManager: No accounts stored.")
	}
	for k, v := range am.Accounts {
		log.Printf("AccountManager: Account: %s [%s]", v.BattleTag, k.String())
	}
}
