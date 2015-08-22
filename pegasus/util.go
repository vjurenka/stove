package pegasus

import (
	"github.com/HearthSim/hs-proto-go/bnet/attribute"
	"github.com/HearthSim/hs-proto-go/bnet/game_utilities_service"
	"github.com/golang/protobuf/proto"
)

// EncodeUtilResponse builds a buffer encoding the response packetId and the
// protobuf message.
func EncodeUtilResponse(packetId int32, msg proto.Message) ([]byte, error) {
	body, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	res := game_utilities_service.ClientResponse{}
	res.Attribute = make([]*attribute.Attribute, 2)
	res.Attribute[0] = &attribute.Attribute{
		Name: proto.String("t"),
		Value: &attribute.Variant{
			IntValue: proto.Int64(int64(packetId)),
		},
	}
	res.Attribute[1] = &attribute.Attribute{
		Name: proto.String("p"),
		Value: &attribute.Variant{
			BlobValue: body,
		},
	}
	return proto.Marshal(&res)
}

func encodeUtilPacketId(systemId, packetId int) int {
	return systemId<<16 | packetId
}
