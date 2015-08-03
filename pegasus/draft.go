package pegasus

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
)

type Draft struct{}

func (s *Draft) Init(sess *Session) {
	sess.RegisterUtilHandler(0, 244, OnDraftGetPicksAndContents)
}

func OnDraftGetPicksAndContents(s *Session, body []byte) ([]byte, error) {
	req := hsproto.PegasusUtil_DraftGetPicksAndContents{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}

	heroDef := MakeCardDef(834, 0)

	res := hsproto.PegasusUtil_DraftChoicesAndContents{
		DeckId:  proto.Int64(0),
		Slot:    proto.Int32(0),
		Wins:    proto.Int32(12),
		Losses:  proto.Int32(0),
		HeroDef: heroDef,
	}

	// stub
	return EncodeUtilResponse(248, &res)
}
