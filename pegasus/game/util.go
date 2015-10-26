package game

import (
	"log"
	"reflect"
)

// Grabs a packetId from proto data, given a proto PacketID enum value.
func packetIDFromProto(x interface{}) int {
	// TODO: de-duplicate between this and pegasus/util.go
	enumTy := reflect.TypeOf(x)
	if enumTy.Kind() != reflect.Ptr {
		enumTy = reflect.PtrTo(enumTy)
	}
	unmarshal, ok := enumTy.MethodByName("UnmarshalJSON")
	if !ok {
		log.Panicf("couldn't get UnmarshalJSON from type: %v", enumTy)
	}
	idV := reflect.New(enumTy.Elem())
	idErr := unmarshal.Func.Call([]reflect.Value{
		idV,
		reflect.ValueOf([]byte(`"ID"`)),
	})[0].Interface()
	if idErr != nil {
		panic(idErr.(error))
	}
	return int(idV.Elem().Int())
}
