package pegasus

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
)

type Version struct{}

func (v *Version) Init(sess *Session) {
	sess.RegisterUtilHandler(0, 303, OnAssetsVersion)
}

func OnAssetsVersion(s *Session, body []byte) ([]byte, error) {
	res := hsproto.PegasusUtil_AssetsVersionResponse{}
	res.Version = proto.Int32(9166)
	return EncodeUtilResponse(304, &res)
}
