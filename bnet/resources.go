package bnet

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"log"
)

type ResourcesServiceBinder struct{}

func (ResourcesServiceBinder) Bind(sess *Session) Service {
	res := &ResourcesService{}
	res.sess = sess
	return res
}

// The Resources service allows the client to query content handles for files
// that are usually often updated.
type ResourcesService struct {
	sess *Session
}

func (s *ResourcesService) Name() string {
	return "bnet.protocol.resources.Resources"
}

func (s *ResourcesService) Methods() []string {
	return []string{
		"",
		"GetContentHandle",
	}
}

func (s *ResourcesService) Invoke(method int, body []byte) (resp []byte, err error) {
	switch method {
	case 1:
		return s.GetContentHandle(body)
	default:
		log.Panicf("error: ResourcesService.Invoke: unknown method %v", method)
		return
	}
}

func (s *ResourcesService) GetContentHandle(body []byte) ([]byte, error) {
	req := hsproto.BnetProtocolResources_ContentHandleRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	res := hsproto.BnetProtocol_ContentHandle{}
	res.Region = proto.Uint32(0x7573)
	res.Usage = proto.Uint32(0x70667479) // 'pfty'
	res.Hash = make([]byte, 0x20)
	resBuf, err := proto.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return resBuf, nil
}
