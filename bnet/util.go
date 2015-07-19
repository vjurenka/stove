package bnet

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"hash/fnv"
)

// EntityId creates a 128-bit entity id from its high and low qwords.
func EntityId(high, low uint64) *hsproto.BnetProtocol_EntityId {
	return &hsproto.BnetProtocol_EntityId{
		High: proto.Uint64(high),
		Low:  proto.Uint64(low),
	}
}

// Hash returns the fnv32a hash of the string.
func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// MakePacket returns a buffer with the encoded values of the supplied header
// and body.
func MakePacket(header *hsproto.BnetProtocol_Header, buf []byte) ([]byte, error) {
	headerBuf, err := proto.Marshal(header)
	if err != nil {
		return nil, err
	}
	headerLen := len(headerBuf)
	packet := make([]byte, 2+headerLen+len(buf))
	packet[0] = byte(headerLen >> 8)
	packet[1] = byte(headerLen & 0xff)
	copy(packet[2:], headerBuf)
	if len(buf) > 0 {
		copy(packet[2+headerLen:], buf)
	}
	return packet, nil
}
