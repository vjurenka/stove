package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/account_service"
	"github.com/HearthSim/hs-proto-go/bnet/account_types"
	"github.com/golang/protobuf/proto"
	"log"
)

type AccountServiceBinder struct{}

func (AccountServiceBinder) Bind(sess *Session) Service {
	res := &AccountService{}
	res.sess = sess
	return res
}

// The Account service holds data about the client's game account.
type AccountService struct {
	sess *Session
}

func (s *AccountService) Name() string {
	return "bnet.protocol.account.AccountService"
}

func (s *AccountService) Methods() []string {
	res := make([]string, 37)
	res[12] = "GetGameAccount"
	res[13] = "GetAccount"
	res[14] = "CreateGameAccount"
	res[15] = "IsIgrAddress"
	res[20] = "CacheExpire"
	res[21] = "CredentialUpdate"
	res[22] = "FlagUpdate"
	res[23] = "GetWalletList"
	res[24] = "GetEBalance"
	res[25] = "Subscribe"
	res[26] = "Unsubscribe"
	res[27] = "GetEBalanceRestrictions"
	res[30] = "GetAccountState"
	res[31] = "GetGameAccountState"
	res[32] = "GetLicenses"
	res[33] = "GetGameTimeRemainingInfo"
	res[34] = "GetGameSessionInfo"
	res[35] = "GetCAISInfo"
	res[36] = "ForwardCacheExpire"
	return res
}

func (s *AccountService) Invoke(method int, body []byte) (resp []byte, err error) {
	switch method {
	case 12:
		return s.GetGameAccount(body)
	case 13:
		return s.GetAccount(body)
	case 14:
		return s.CreateGameAccount(body)
	case 15:
		return []byte{}, s.IsIgrAddress(body)
	case 20:
		return nil, s.CacheExpire(body)
	case 21:
		return nil, s.CredentialUpdate(body)
	case 22:
		return s.FlagUpdate(body)
	case 23:
		return s.GetWalletList(body)
	case 24:
		return s.GetEBalance(body)
	case 25:
		return s.Subscribe(body)
	case 26:
		return []byte{}, s.Unsubscribe(body)
	case 27:
		return s.GetEBalanceRestrictions(body)
	case 30:
		return s.GetAccountState(body)
	case 31:
		return s.GetGameAccountState(body)
	case 32:
		return s.GetLicenses(body)
	case 33:
		return s.GetGameTimeRemainingInfo(body)
	case 34:
		return s.GetGameSessionInfo(body)
	case 35:
		return s.GetCAISInfo(body)
	case 36:
		return []byte{}, s.ForwardCacheExpire(body)
	default:
		log.Panicf("error: AccountService.Invoke: unknown method %v", method)
		return
	}
}

func (s *AccountService) GetGameAccount(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) GetAccount(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) CreateGameAccount(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) IsIgrAddress(body []byte) error {
	return nyi
}

func (s *AccountService) CacheExpire(body []byte) error {
	return nyi
}

func (s *AccountService) CredentialUpdate(body []byte) error {
	return nyi
}

func (s *AccountService) FlagUpdate(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) GetWalletList(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) GetEBalance(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) Subscribe(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) Unsubscribe(body []byte) error {
	return nyi
}

func (s *AccountService) GetEBalanceRestrictions(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) GetAccountState(body []byte) ([]byte, error) {
	req := account_service.GetAccountStateRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	opts := req.GetOptions()
	res := account_service.GetAccountStateResponse{
		State: &account_types.AccountState{},
	}
	if opts.GetAllFields() || opts.GetFieldAccountLevelInfo() {
		levelInfo := &account_types.AccountLevelInfo{}
		levelInfo.PreferredRegion = proto.Uint32(1) // US
		levelInfo.Country = proto.String("United States")
		levelInfo.Licenses = make([]*account_types.AccountLicense, 1)
		levelInfo.Licenses[0] = &account_types.AccountLicense{
			Id: proto.Uint32(1),
		}
		res.State.AccountLevelInfo = levelInfo
	}
	return proto.Marshal(&res)
}

func (s *AccountService) GetGameAccountState(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) GetLicenses(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) GetGameTimeRemainingInfo(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) GetGameSessionInfo(body []byte) ([]byte, error) {
	req := account_service.GetGameSessionInfoRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	res := account_service.GetGameSessionInfoResponse{}
	res.SessionInfo = &account_types.GameSessionInfo{}
	res.SessionInfo.StartTime = proto.Uint32(uint32(s.sess.startedPlaying.Unix()))
	return proto.Marshal(&res)
}

func (s *AccountService) GetCAISInfo(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AccountService) ForwardCacheExpire(body []byte) error {
	return nyi
}
