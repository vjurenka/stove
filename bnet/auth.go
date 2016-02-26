package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/authentication_service"
	"github.com/HearthSim/hs-proto-go/bnet/entity"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

type AuthServerServiceBinder struct{}

func (AuthServerServiceBinder) Bind(sess *Session) Service {
	res := &AuthServerService{}
	res.sess = sess
	return res
}

// The AuthServer service handles Logon requests.  This implementation does not
// use the Module system but instead relies on a pre-shared WebAuth token to
// authenticate the client.
type AuthServerService struct {
	sess *Session

	program  string
	email    string
	loggedIn bool
	battleTag string
	client   *AuthClientService
	ent_id uint64
}

func (s *AuthServerService) Name() string {
	return "bnet.protocol.authentication.AuthenticationServer"
}

func (s *AuthServerService) Methods() []string {
	return []string{
		"",
		"Logon",
		"ModuleNotify",
		"ModuleMessage",
		"SelectGameAccount_DEPRECATED",
		"GenerateTempCookie",
		"SelectGameAccount",
		"VerifyWebCredentials",
	}
}

func (s *AuthServerService) Invoke(method int, body []byte) (resp []byte, err error) {
	switch method {
	case 1:
		return []byte{}, s.Logon(body)
	case 2:
		return []byte{}, s.ModuleNotify(body)
	case 3:
		return []byte{}, s.ModuleMessage(body)
	case 4:
		return []byte{}, s.SelectGameAccount_DEPRECATED(body)
	case 5:
		return s.GenerateTempCookie(body)
	case 6:
		return []byte{}, s.SelectGameAccount(body)
	case 7:
		return []byte{}, s.VerifyWebCredentials(body)
	default:
		log.Panicf("error: AuthServerService.Invoke: unknown method %v", method)
		return
	}
}

func (s *AuthServerService) Logon(body []byte) error {
	req := authentication_service.LogonRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())
	s.program = req.GetProgram()
	// TODO: pull account from db
	log.Printf("logon request from %s", req.GetEmail())
	s.email = string(req.GetEmail())
	s.client = s.sess.ImportedService("bnet.protocol.authentication.AuthenticationClient").(*AuthClientService)
	s.FinishQueue()
	s.sess.Transition(StateLoggingIn)
	return nil
}

func (s *AuthServerService) ModuleNotify(body []byte) error {
	return nyi
}

func (s *AuthServerService) ModuleMessage(body []byte) error {
	return nyi
}

func (s *AuthServerService) SelectGameAccount_DEPRECATED(body []byte) error {
	req := entity.EntityId{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())
	return nil
}

func (s *AuthServerService) GenerateTempCookie(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AuthServerService) SelectGameAccount(body []byte) error {
	return nyi
}

func (s *AuthServerService) VerifyWebCredentials(body []byte) error {
	req := authentication_service.VerifyWebCredentialsRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())
	account := []Account{}
	s.loggedIn = false
	db.Where("email = ? and web_credential = ?", string(s.email), string(req.GetWebCredentials())).First(&account)
	if len(account) != 0 {
		log.Printf("account %s (BattleTag: %s) authorized", account[0].Email, account[0].BattleTag)
		s.loggedIn = true
		s.battleTag = account[0].BattleTag
		s.ent_id = account[0].ID
	}
	return s.CompleteLogin()
}

func (s *AuthServerService) CompleteLogin() error {
	res := authentication_service.LogonResult{}
	if !s.loggedIn {
		res.ErrorCode = proto.Uint32(ErrorNoAuth)
	} else {
		res.ErrorCode = proto.Uint32(ErrorOK)
		// TODO: Make this data real.  ConnectGameServer needs to return the
		// GameAccount EntityId.
		//res.Account = EntityId(72058118023938048, 1)
		res.Account = EntityId(72057594037927936+s.ent_id, 1)
		s.sess.account = s.sess.server.accountManager.AddAccount(res.Account.GetHigh(), res.Account.GetLow(), s.battleTag, s.sess)
		s.sess.server.accountManager.Dump()
		res.GameAccount = make([]*entity.EntityId, 1)
		//res.GameAccount[0] = EntityId(144115713527006023, 1)
		res.GameAccount[0] = EntityId(144115188075855872+s.ent_id, 2)
		s.sess.server.accountManager.AddGameAccount(res.GameAccount[0].GetHigh(), res.GameAccount[0].GetLow())
		res.ConnectedRegion = proto.Uint32(0x5553) // 'US'

		if s.program == "WTCG" {
			s.sess.server.ConnectGameServer(s.sess, s.program)
			go s.sess.HandleNotifications()
		}
		s.sess.startedPlaying = time.Now()
	}
	resBody, err := proto.Marshal(&res)
	if err != nil {
		return err
	}
	resHeader := s.sess.MakeRequestHeader(s.client, 5, len(resBody))
	err = s.sess.QueuePacket(resHeader, resBody)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthServerService) FinishQueue() {
	update := authentication_service.LogonQueueUpdateRequest{}
	update.Position = proto.Uint32(0)
	update.EstimatedTime = proto.Uint64(0)
	update.EtaDeviationInSec = proto.Uint64(0)
	updateBody, err := proto.Marshal(&update)
	if err != nil {
		log.Panicf("FinishQueue: %v", err)
	}
	updateHeader := s.sess.MakeRequestHeader(s.client, 12, len(updateBody))
	s.sess.QueuePacket(updateHeader, updateBody)

	endHeader := s.sess.MakeRequestHeader(s.client, 13, 0)
	s.sess.QueuePacket(endHeader, nil)
}

type AuthClientServiceBinder struct{}

func (AuthClientServiceBinder) Bind(sess *Session) Service {
	service := &AuthClientService{sess}
	return service
}

// The AuthClient service is implemented by the client to handle auth modules
// and Logon results.
type AuthClientService struct {
	sess *Session
}

func (s *AuthClientService) Name() string {
	return "bnet.protocol.authentication.AuthenticationClient"
}

func (s *AuthClientService) Methods() []string {
	res := make([]string, 15)
	res[1] = "ModuleLoad"
	res[2] = "ModuleMessage"
	res[3] = "AccountSettings"
	res[4] = "ServerStateChange"
	res[5] = "LogonComplete"
	res[6] = "MemModuleLoad"
	res[10] = "LogonUpdate"
	res[11] = "VersionInfoUpdated"
	res[12] = "LogonQueueUpdate"
	res[13] = "LogonQueueEnd"
	res[14] = "GameAccountSelected"
	return res
}

func (s *AuthClientService) Invoke(method int, body []byte) (resp []byte, err error) {
	log.Panicf("AuthClientService is a client export, not a server export")
	return
}
