package pegasus

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
)

type Store struct{}

func (s *Store) Init(sess *Session) {
	sess.RegisterUtilHandler(1, 237, OnGetBattlePayConfig)
	sess.RegisterUtilHandler(1, 255, OnGetBattlePayStatus)
}

func OnGetBattlePayConfig(s *Session, body []byte) ([]byte, error) {
	res := hsproto.PegasusUtil_BattlePayConfigResponse{}
	return EncodeUtilResponse(238, &res)
}

func OnGetBattlePayStatus(s *Session, body []byte) ([]byte, error) {
	res := hsproto.PegasusUtil_BattlePayStatusResponse{}
	status := hsproto.PegasusUtil_BattlePayStatusResponse_PS_READY
	res.Status = &status
	res.BattlePayAvailable = proto.Bool(false)
	return EncodeUtilResponse(265, &res)
}
