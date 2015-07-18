package bnet

import (
	"fmt"
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"hash/fnv"
	"log"
	"net"
	"time"
)

// A Service is a set of RPC methods bound to a particular Session.
type Service interface {
	// Name returns the fully qualified name of the service.
	Name() string

	// Methods returns the names of the service's invokable methods.
	Methods() []string

	// Invoke executes a method.
	Invoke(method int, body []byte) (resp []byte, err error)
}

var nyi = fmt.Errorf("nyi")

// Hash returns the hash by which the service is described.
func ServiceHash(s Service) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s.Name()))
	return h.Sum32()
}

type Server struct {
	registeredServices []Service
}

func NewServer() *Server {
	s := &Server{}
	return s
}

func (s *Server) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handleClient(c)
	}
}

func (s *Server) handleClient(c net.Conn) {
	c.SetDeadline(time.Time{})

	sess := NewSession(s, c)
	// The connection service is implicity exported:
	sess.Export(0, NewConnectionService(sess))

	buf := make([]byte, 0x1000)
	for {
		_, err := sess.conn.Read(buf[:2])
		if err != nil {
			log.Panicf("error: Server.handleClient: length read: %v", err)
		}
		headerLen := int(buf[0])<<8 | int(buf[1])
		if headerLen > len(buf) {
			buf = append(buf, make([]byte, headerLen-len(buf))...)
		}
		_, err = sess.conn.Read(buf[:headerLen])
		if err != nil {
			log.Panicf("error: Server.handleClient: header read: %v", err)
		}
		header := hsproto.BnetProtocol_Header{}
		err = proto.Unmarshal(buf[:headerLen], &header)
		if err != nil {
			log.Panicf("error: Server.handleClient: header decode: %v", err)
		}
		bodyLen := int(header.GetSize())
		if bodyLen > len(buf) {
			buf = append(buf, make([]byte, bodyLen-len(buf))...)
		}
		_, err = sess.conn.Read(buf[:bodyLen])
		if err != nil {
			log.Panicf("error: Server.handleClient: body read: %v", err)
		}
		log.Printf("handling packet %s %x", header.String(), buf[:bodyLen])
		sess.HandlePacket(&header, buf[:bodyLen])
	}
}

type Session struct {
	server *Server
	conn   net.Conn

	// Exports contain methods the client may invoke on the server; the client
	// refers to these as imports.
	exports []Service
	// Imports contain methods the server may invoke on the client; the client
	// refers to these as exports.
	imports []Service

	// A request registers itself for a response by assigning to this map a
	// channel on which it will listen for the response.
	responses map[uint32]chan []byte
	// The token used for request/response pairs increments sequentially.
	lastToken uint32
}

func NewSession(s *Server, c net.Conn) *Session {
	sess := &Session{}
	sess.server = s
	sess.conn = c
	sess.responses = map[uint32]chan []byte{}
	return sess
}

func (s *Session) Export(index int, service Service) {
	if index >= len(s.exports) {
		padLen := (1 + index) - len(s.exports)
		s.exports = append(s.exports, make([]Service, padLen)...)
	}
	s.exports[index] = service
}

func (s *Session) Import(index int, service Service) {
	if index >= len(s.imports) {
		padLen := (1 + index) - len(s.imports)
		s.imports = append(s.imports, make([]Service, padLen)...)
	}
	s.imports[index] = service
}

func (s *Session) Write(header *hsproto.BnetProtocol_Header, buf []byte) error {
	headerBuf, err := proto.Marshal(header)
	if err != nil {
		return err
	}
	headerLen := len(headerBuf)
	packet := make([]byte, 2+headerLen+len(buf))
	packet[0] = byte(headerLen >> 8)
	packet[1] = byte(headerLen & 0xff)
	copy(packet[2:], headerBuf)
	copy(packet[2+headerLen:], buf)
	_, err = s.conn.Write(packet)
	log.Printf("Wrote %d bytes", len(packet))
	if err != nil {
		return err
	}
	return nil
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
				Token:     proto.Uint32(header.GetToken()),
				Size:      proto.Uint32(uint32(len(resp))),
			}
			err := s.Write(&respHead, resp)
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
	resp, err := service.Invoke(methodId, body)
	if err != nil {
		log.Panicf("error: Session.HandleRequest: Invoke: %v", err)
	}
	return resp
}
