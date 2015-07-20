package pegasus

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"log"
)

type Account struct{}

func (v *Account) Init(sess *Session) {
	sess.RegisterUtilHandler(0, 267, OnCheckAccountLicenses)
	sess.RegisterUtilHandler(1, 276, OnCheckGameLicenses)
	sess.RegisterUtilHandler(0, 205, OnUpdateLogin)
}

func OnCheckAccountLicenses(s *Session, body []byte) ([]byte, error) {
	return OnCheckLicenses(true)
}

func OnCheckGameLicenses(s *Session, body []byte) ([]byte, error) {
	return OnCheckLicenses(false)
}

func OnCheckLicenses(accountLevel bool) ([]byte, error) {
	res := hsproto.PegasusUtil_CheckLicensesResponse{}
	res.AccountLevel = proto.Bool(accountLevel)
	res.Success = proto.Bool(true)
	return EncodeUtilResponse(277, &res)
}

func OnUpdateLogin(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_UpdateLogin{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	res := hsproto.PegasusUtil_UpdateLoginComplete{}
	return EncodeUtilResponse(307, &res)
}
