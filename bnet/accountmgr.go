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

func (e *amEntityId) String() string { return "["+string(e.High)+","+string(e.Low)+"]" }

func amEntityId_convert(source *entity.EntityId) *amEntityId {
	if source != nil {
		return &amEntityId{source.GetHigh(), source.GetLow()}
	}
	return nil
}

type amPresenceKey struct {
	High uint64
	Low uint64
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
	Email string
	Session *Session
}

type amGameAccount struct {
	EntityId amEntityId
}

type AccountManager struct {
	Accounts map[amEntityId]*amAccount
	GameAccounts map[amEntityId]*amGameAccount
	BattleTags map[string]amEntityId
	PresenceData map[amPresenceKey]amPresenceData
	Subscribers map[amEntityId][]*amAccount
}

func NewAccountManager() *AccountManager {
	log.Print("AccountManager: initializing")
	return &AccountManager{
		Accounts: map[amEntityId]*amAccount{},
		GameAccounts: map[amEntityId]*amGameAccount{},
		BattleTags: map[string]amEntityId{},
		PresenceData: map[amPresenceKey]amPresenceData{},
		Subscribers: map[amEntityId][]*amAccount{},
	}
}

func (am *AccountManager) GetBattleTag(e amEntityId) string {
	v, ok := am.Accounts[e]
	if !ok {
		return "Unknown"
	}
	return v.BattleTag
}

func (am *AccountManager) AddAccount(high uint64, low uint64, battletag string, email string, session *Session) *amAccount {
	/*_, ok := am.Accounts[amEntityId{high, low}]
	if !ok {
		log.Printf("AccountManager: AddAccount: account already created")
	}*/
	// check low == 1 (Player account)
	/*if low != 1 {
		log.Printf("AccountManager: Cannot add non-account class entry")
		return nil
	}*/
	// account doesn't exist yet
	res := amAccount{amEntityId{high, low}, battletag, email, session}
	log.Printf("AccountManager: AddAccount: %+v", res)
	am.Accounts[amEntityId{high, low}] = &res
	am.BattleTags[battletag] = amEntityId{high, low}
	return &res
}

func (am *AccountManager) AddGameAccount(high uint64, low uint64) bool {
	_, ok := am.GameAccounts[amEntityId{high, low}]
	if !ok {
		// check low == 2 (Game account)
		/*if low != 2 {
			log.Printf("AccountManager: Cannot add non-game class entry")
			return false
		}*/
		// account doesn't exist yet
		am.GameAccounts[amEntityId{high, low}] = &amGameAccount{amEntityId{high, low}}
		return true
	}
	return false
}

func (am *AccountManager) GetPresenceData(k amPresenceKey) attribute.Variant {
	log.Printf("AccountManager: GetPresenceData: [%+v]", k)
	presence, ok := am.PresenceData[k]
	if !ok {
		return attribute.Variant{}
	}
	return presence.data
}

func (am *AccountManager) UpdatePresenceData(k amPresenceKey, v attribute.Variant) bool {
	log.Printf("AccountManager: UpdatePresenceData: [%+v] = %+v", k, v)
	presence, ok := am.PresenceData[k]
	if !ok {
		am.PresenceData[k] = amPresenceData{}
	}
	presence.data = v
	presence.update_time = time.Now()
	am.PresenceData[k] = presence
	return true
}
func (am *AccountManager) RemovePresenceData(k amPresenceKey) bool {
	log.Printf("AccountManager: RemovePresenceData: [%+v]", k)
	delete(am.PresenceData, k)
	return true
}

func (am *AccountManager) Subscribe(s amEntityId, d *amAccount) bool {
	_, ok := am.Subscribers[s]
	if !ok {
		log.Printf("AccountManager: Subscriber list [%+v] not present. Creating new one.", s)
		//return false
		am.Subscribers[s] = []*amAccount{}
	}
	am.Subscribers[s] = append(am.Subscribers[s], d)
	log.Printf("AccountManager: Account [%+v] will be notified about changes in [%+v].", d, s)
	return true
}

func (am AccountManager) Dump() {
	if len(am.Accounts) == 0{
		log.Printf("AccountManager: No accounts stored.")
	}
	for k, v := range am.Accounts {
		log.Printf("AccountManager: Account: %s [%+v]", v.BattleTag, k)
	}
}
