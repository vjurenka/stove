package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/entity"
	"github.com/HearthSim/hs-proto-go/bnet/attribute"
	"time"
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

func amEntityId_convert(source *entity.EntityId) *amEntityId {
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

type amPresenceData struct {
	data attribute.Variant
	update_time time.Time
}

type amAccount struct {
	EntityId amEntityId
	BattleTag string
	Session *Session
	Subscribers []*amAccount
	PresenceData map[amPresenceKey]amPresenceData
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

func (am *AccountManager) AddAccount(high uint64, low uint64, battletag string, session *Session) *amAccount {
	_, ok := am.Accounts[amEntityId{high, low}]
	if !ok {
		// check low == 1 (Player account)
		if low != 1 {
			log.Printf("AccountManager: Cannot add non-account class entry")
			return nil
		}
		// account doesn't exist yet
		res := &amAccount{amEntityId{high, low}, battletag, session, []*amAccount{}, map[amPresenceKey]amPresenceData{}}
		am.Accounts[amEntityId{high, low}] = res
		am.BattleTags[battletag] = amEntityId{high, low}
		return res
	}
	return nil
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

func (am *AccountManager) GetPresenceData(e amEntityId, k amPresenceKey) attribute.Variant {
	log.Printf("AccountManager: GetPresenceData: [%s %s]", e, k)
	account, ok := am.Accounts[e]
	if !ok {
		return attribute.Variant{}
	}
	presence, ok := account.PresenceData[k]
	if !ok {
		return attribute.Variant{}
	}
	return presence.data
}

func (am *AccountManager) UpdatePresenceData(e amEntityId, k amPresenceKey, v attribute.Variant) bool {
	log.Printf("AccountManager: UpdatePresenceData: [%s %s] = %s", e, k, v)
	account, ok := am.Accounts[e]
	if !ok {
		// TODO: check if we can safely ignore this, or if it is needed to accept data for non-accounts
		log.Panicf("AccountManager: Account is not registered")
	}
	presence, ok := account.PresenceData[k]
	if !ok {
		account.PresenceData[k] = amPresenceData{}
	}
	presence.data = v
	presence.update_time = time.Now()
	return true
}
func (am *AccountManager) RemovePresenceData(e amEntityId, k amPresenceKey) bool {
	log.Printf("AccountManager: RemovePresenceData: [%s %s]", e, k)
	delete(am.Accounts[e].PresenceData, k)
	return true
}

func (am *AccountManager) Subscribe(s amEntityId, d *amAccount) bool {
	return true
}

func (am AccountManager) Dump() {
	if len(am.Accounts) == 0{
		log.Printf("AccountManager: No accounts stored.")
	}
	for k, v := range am.Accounts {
		log.Printf("AccountManager: Account: %s [%s]", v.BattleTag, k.String())
	}
}
