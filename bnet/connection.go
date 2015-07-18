package bnet

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

type ConnectionService struct {
	sess *Session
}

type ConnectionServiceBinder struct{}

func (ConnectionServiceBinder) Bind(sess *Session) Service {
	return &ConnectionService{sess}
}

func (s *ConnectionService) Name() string {
	return "bnet.protocol.connection.ConnectionService"
}

func (s *ConnectionService) Methods() []string {
	return []string{
		"",
		"Connect",
		"Bind",
		"Echo",
		"ForceDisconnect",
		"KeepAlive",
		"Encrypt",
		"RequestDisconnect",
	}
}

func (s *ConnectionService) Invoke(method int, body []byte) (resp []byte, err error) {
	switch method {
	case 1:
		return s.Connect(body)
	case 2:
		return s.Bind(body)
	case 3:
		return s.Echo(body)
	case 4:
		return nil, s.ForceDisconnect(body)
	case 5:
		return nil, s.KeepAlive(body)
	case 6:
		return []byte{}, s.Encrypt(body)
	case 7:
		return nil, s.RequestDisconnect(body)
	default:
		log.Panicf("error: ConnectionService.Invoke: unknown method %v", method)
		return
	}
}

func (s *ConnectionService) Connect(body []byte) ([]byte, error) {
	req := hsproto.BnetProtocolConnection_ConnectRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Println("req:", req)
	bindReq := req.GetBindRequest()
	serviceId := len(s.sess.exports)
	exportedServiceIds := []uint32{}
	for _, exportRequest := range bindReq.GetImportedServiceHash() {
		exportedServiceIds = append(exportedServiceIds, uint32(serviceId))
		s.sess.BindExport(serviceId, exportRequest)
		serviceId += 1
	}
	now := time.Now()
	nowNano := uint64(now.UnixNano())
	nowSec := uint32(now.Unix())
	resp := hsproto.BnetProtocolConnection_ConnectResponse{
		ServerId: &hsproto.BnetProtocol_ProcessId{
			Label: proto.Uint32(3868510373),
			Epoch: proto.Uint32(nowSec),
		},
		ClientId: &hsproto.BnetProtocol_ProcessId{
			Label: proto.Uint32(1255760),
			Epoch: proto.Uint32(nowSec),
		},
		BindResult: proto.Uint32(0),
		BindResponse: &hsproto.BnetProtocolConnection_BindResponse{
			ImportedServiceId: exportedServiceIds,
		},
		ServerTime: proto.Uint64(nowNano),
	}
	log.Println("resp:", resp)
	respBuf, err := proto.Marshal(&resp)
	if err != nil {
		return nil, err
	}
	return respBuf, nil
}

func (s *ConnectionService) Bind(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *ConnectionService) Echo(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *ConnectionService) ForceDisconnect(body []byte) error {
	return nyi
}

func (s *ConnectionService) KeepAlive(body []byte) error {
	return nyi
}

func (s *ConnectionService) Encrypt(body []byte) error {
	return nyi
}

func (s *ConnectionService) RequestDisconnect(body []byte) error {
	return nyi
}
