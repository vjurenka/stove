package pegasus

import (
	"github.com/golang/protobuf/proto"
	"log"
	"reflect"
)

type Packet struct {
	Body   []byte
	ID     int32
	System int32
}

type PacketID struct {
	ID     int32
	System int32
}

type UtilHandler func(sess *Session, req []byte) *Packet

// EncodePacket builds a Packet with the protobuf message and packet id.
func EncodePacket(x interface{}, msg proto.Message) *Packet {
	body, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	id := packetIDFromProto(x)
	return &Packet{body, id.ID, id.System}
}

// Grabs a packetId from proto data, given a proto PacketID enum value.
func packetIDFromProto(x interface{}) PacketID {
	res := PacketID{}
	enumTy := reflect.TypeOf(x)
	if enumTy.Kind() != reflect.Ptr {
		enumTy = reflect.PtrTo(enumTy)
	}
	unmarshal, ok := enumTy.MethodByName("UnmarshalJSON")
	if !ok {
		log.Panicf("couldn't get UnmarshalJSON from type: %v", enumTy)
	}
	idV := reflect.New(enumTy.Elem())
	sysV := reflect.New(enumTy.Elem())
	idErr := unmarshal.Func.Call([]reflect.Value{
		idV,
		reflect.ValueOf([]byte(`"ID"`)),
	})[0].Interface()
	if idErr != nil {
		panic(idErr.(error))
	}
	res.ID = int32(idV.Elem().Int())
	sysErr := unmarshal.Func.Call([]reflect.Value{
		sysV,
		reflect.ValueOf([]byte(`"System"`)),
	})[0].Interface()
	if sysErr != nil {
		res.System = 0
	} else {
		res.System = int32(sysV.Elem().Int())
	}
	return res
}
