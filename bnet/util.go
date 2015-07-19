package bnet

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
)

func EntityId(high, low uint64) *hsproto.BnetProtocol_EntityId {
	return &hsproto.BnetProtocol_EntityId{
		High: proto.Uint64(high),
		Low:  proto.Uint64(low),
	}
}
