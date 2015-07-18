package bnet

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"sync"
)

const (
	StateDisconnected = iota
	StateConnecting
	StateConnected
	StateLoggingIn
	StateAuthenticationFailed
	StateReady
)

type Session struct {
	server *Server
	conn   net.Conn

	// Exports contain methods the client may invoke on the server; the client
	// refers to these as imports.
	exports []Service
	// Imports contain methods the server may invoke on the client; the client
	// refers to these as exports.
	imports []Service
	// Maps an imported service hash to an index into imports.
	importMap map[uint32]int

	// A request registers itself for a response by assigning to this map a
	// channel on which it will listen for the response.
	responses map[uint32]chan []byte
	// The token used for request/response pairs increments sequentially.
	lastToken uint32

	// This channel contains outgoing packets.
	packetQueue chan []byte

	StateMutex  sync.Mutex // protects Session.State
	StateChange *sync.Cond
	State       int
}

func NewSession(s *Server, c net.Conn) *Session {
	sess := &Session{}
	sess.server = s
	sess.conn = c
	sess.responses = map[uint32]chan []byte{}
	sess.packetQueue = make(chan []byte)
	sess.StateChange = sync.NewCond(&sess.StateMutex)
	sess.State = StateConnecting
	// The connection service export is implicity bound at index 0:
	sess.BindExport(0, Hash("bnet.protocol.connection.ConnectionService"))
	go sess.pumpPacketQueue()
	return sess
}

func (s *Session) BindExport(index int, hash uint32) {
	var service Service = nil
	binder, ok := s.server.registeredServices[hash]
	if !ok {
		log.Printf("warn: Session.BindExport: unknown service hash: %x", hash)
		// We still want to put a nil in the slot, so that we panic when the
		// service is invoked.
	} else {
		service = binder.Bind(s)
	}
	if index >= len(s.exports) {
		padLen := (1 + index) - len(s.exports)
		s.exports = append(s.exports, make([]Service, padLen)...)
	}
	s.exports[index] = service
}

func (s *Session) BindImport(index int, hash uint32) {
	var service Service = nil
	binder, ok := s.server.registeredServices[hash]
	if !ok {
		log.Printf("warn: Session.BindImport: unknown service hash: %x", hash)
	} else {
		service = binder.Bind(s)
	}
	if index >= len(s.imports) {
		padLen := (1 + index) - len(s.imports)
		s.imports = append(s.imports, make([]Service, padLen)...)
	}
	s.imports[index] = service
}

func (s *Session) QueuePacket(header *hsproto.BnetProtocol_Header, buf []byte) error {
	packet, err := MakePacket(header, buf)
	if err != nil {
		return err
	}
	s.packetQueue <- packet
	return nil
}

// Goroutine to pump the outgoing packet queue
func (s *Session) pumpPacketQueue() {
	for {
		packet := <-s.packetQueue
		_, err := s.conn.Write(packet)
		log.Printf("Wrote %d bytes", len(packet))
		if err != nil {
			log.Panicf("error: Session.WritePacketQueue failed: %v", err)
		}
	}
}

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
	copy(packet[2+headerLen:], buf)
	return packet, nil
}

func (s *Session) MakeRequestHeader(service Service, methodId, size int) *hsproto.BnetProtocol_Header {
	serviceId, ok := s.importMap[Hash(service.Name())]
	if !ok {
		log.Panicf("Client didn't export service %s", service.Name())
	}
	token := s.lastToken
	s.lastToken++
	return &hsproto.BnetProtocol_Header{
		ServiceId: proto.Uint32(uint32(serviceId)),
		MethodId:  proto.Uint32(uint32(methodId)),
		Token:     proto.Uint32(token),
		Size:      proto.Uint32(uint32(size)),
	}
}

func (s *Session) HandlePacket(header *hsproto.BnetProtocol_Header, body []byte) {
	serviceId := int(header.GetServiceId())
	methodId := int(header.GetMethodId())

	if serviceId == 254 {
		s.HandleResponse(header.GetToken(), body)
	} else {
		resp := s.HandleRequest(serviceId, methodId, body)
		if resp != nil {
			respHead := hsproto.BnetProtocol_Header{
				ServiceId: proto.Uint32(254),
				MethodId:  proto.Uint32(0),
				Token:     header.Token,
				Size:      proto.Uint32(uint32(len(resp))),
			}
			err := s.QueuePacket(&respHead, resp)
			if err != nil {
				log.Panicf("error: Session.HandlePacket: respond: %v", err)
			}
		}
	}
}

func (s *Session) HandleResponse(token uint32, body []byte) {
	if ch, ok := s.responses[token]; ok {
		// Note: don't use unbuffered channels for response channels, please...
		ch <- body
	} else {
		log.Printf(" warn: Session.HandleResponse: token not found: %v", token)
	}
}

func (s *Session) HandleRequest(serviceId, methodId int, body []byte) (resp []byte) {
	var service Service
	if serviceId < len(s.exports) {
		service = s.exports[serviceId]
	}
	if service == nil {
		log.Panicf("error: Session.HandleRequest: Unknown serviceId %v", serviceId)
	}
	serviceName := service.Name()
	methodNames := service.Methods()
	methodName := "(unknown)"
	if len(methodNames) > methodId {
		methodName = methodNames[methodId]
	}
	log.Printf("Session.HandleRequest: invoking %s.%s", serviceName, methodName)
	resp, err := service.Invoke(methodId, body)
	if err != nil {
		log.Panicf("error: Session.HandleRequest: Invoke: %v", err)
	}
	return resp
}

func (s *Session) Transition(state int) {
	s.StateMutex.Lock()
	currState := s.State
	s.State = state
	s.StateMutex.Unlock()
	if state != currState {
		s.StateChange.Broadcast()
	}
}

func (s *Session) WaitForTransition(state int) {
	s.StateMutex.Lock()
	defer s.StateMutex.Unlock()
	for s.State != state {
		s.StateChange.Wait()
	}
}
