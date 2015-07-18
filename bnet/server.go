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

// Use nyi to error from unimplemented service methods.
var nyi = fmt.Errorf("nyi")

// A Service is a set of RPC methods bound to a particular Session.
type Service interface {
	// Name returns the fully qualified name of the service.
	Name() string

	// Methods returns the names of the service's invokable methods.
	Methods() []string

	// Invoke executes a method.
	Invoke(method int, body []byte) (resp []byte, err error)
}

type ServiceBinder interface {
	// Bind binds a service to a session.  Passing nil will give a default
	// instance which can be used to inspect the service and method names.
	Bind(sess *Session) Service
}

// Hash returns the fnv32a hash of the string.
func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// ServiceHash returns a hash of the service's fully qualified name.
func ServiceHash(binder ServiceBinder) uint32 {
	s := binder.Bind(nil)
	return Hash(s.Name())
}

type Server struct {
	// Registered services are mapped by their service hash to a service binder.
	registeredServices map[uint32]ServiceBinder
}

func NewServer() *Server {
	s := &Server{}
	s.registeredServices = map[uint32]ServiceBinder{}
	s.registerService(ConnectionServiceBinder{})
	s.registerService(AuthServerServiceBinder{})
	s.registerService(ChallengeNotifyServiceBinder{})
	return s
}

func (s *Server) registerService(binder ServiceBinder) {
	s.registeredServices[ServiceHash(binder)] = binder
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
