package pegasus

import (
	"errors"
	"github.com/HearthSim/hs-proto/go"
	"github.com/HearthSim/stove/bnet"
	"log"
)

type Session struct {
	server   *Server
	host     *bnet.Session
	handlers map[int]UtilHandler

	Account
	Draft
	Store
	Subscription
	Version
}

func NewSession(s *Server, hostSess *bnet.Session) *Session {
	sess := &Session{}
	sess.server = s
	sess.host = hostSess
	sess.handlers = map[int]UtilHandler{}

	sess.Account.Init(sess)
	sess.Draft.Init(sess)
	sess.Store.Init(sess)
	sess.Subscription.Init(sess)
	sess.Version.Init(sess)
	return sess
}

var nyi = errors.New("not yet implemented")

type UtilHandler func(sess *Session, req []byte) (res []byte, err error)

// HandleUtilRequest processes an encoded client request from GameUtilities,
// possibly returning an encoded response.
func (s *Session) HandleUtilRequest(req *hsproto.BnetProtocolGameUtilities_ClientRequest) ([]byte, error) {
	packetId := -1
	// System 1 involves payment.  System 0 is everything else.
	systemId := -1
	route := uint64(0)
	var data []byte
	for _, attr := range req.Attribute {
		key := attr.GetName()
		val := attr.GetValue()
		switch key {
		case "p":
			blob := val.GetBlobValue()
			if len(blob) < 2 {
				log.Panicf("bad util packet: %s", req.String())
			}
			packetId = int(blob[0]) | int(blob[1])<<8
			data = blob[2:]
		case "v":
			if val.IntValue != nil {
				intVal := int(val.GetIntValue())
				systemId = intVal % 10
			} else if val.StringValue != nil {
				strVal := val.GetStringValue()
				sysStr := strVal[len(strVal)-1:]
				switch sysStr {
				case "b": // BattlePay
					systemId = 1
				case "c": // ConnectAPI
					systemId = 0
				default:
					log.Panicf("bad util packet: %s", req.String())
				}
			} else {
				log.Panicf("bad util packet: %s", req.String())
			}
		case "r":
			route = val.GetUintValue()
		default:
			log.Printf("HandleUtilRequest: unknown key: %s = %s", key, val.String())
		}
	}
	if route != s.route {
		log.Printf("HandleUtilRequest: bad route")
	}
	return s.handleUtilRequest(systemId, packetId, data)
}

func (s *Session) handleUtilRequest(systemId, packetId int, req []byte) ([]byte, error) {
	id := encodeUtilPacketId(systemId, packetId)
	if handler, ok := s.handlers[id]; ok {
		return handler(s, req)
	} else {
		log.Panicf("handler does not exist for util packet %d:%d", systemId, packetId)
	}
	return nil, nyi
}

func (s *Session) RegisterUtilHandler(systemId, packetId int, handler UtilHandler) {
	id := encodeUtilPacketId(systemId, packetId)
	if _, ok := s.handlers[id]; !ok {
		s.handlers[id] = handler
	} else {
		log.Panicf("cannot overwrite existing handler for util packet %d:%d", systemId, packetId)
	}
}

func (s *Session) UnregisterUtilHandler(systemId, packetId int, handler UtilHandler) {
	id := encodeUtilPacketId(systemId, packetId)
	if _, ok := s.handlers[id]; ok {
		delete(s.handlers, id)
	} else {
		log.Panicf("unregister called for non-existent handler for util packet %d:%d", systemId, packetId)
	}
}
