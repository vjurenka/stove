package bnet

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"log"
)

type AuthServerServiceBinder struct{}

func (AuthServerServiceBinder) Bind(sess *Session) Service {
	return &AuthServerService{sess}
}

type AuthServerService struct {
	sess *Session
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
	req := hsproto.BnetProtocolAuthentication_LogonRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())
	log.Printf("logon request from %s", req.GetEmail())
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
	return nyi
}

func (s *AuthServerService) GenerateTempCookie(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *AuthServerService) SelectGameAccount(body []byte) error {
	return nyi
}

func (s *AuthServerService) VerifyWebCredentials(body []byte) error {
	return nyi
}
