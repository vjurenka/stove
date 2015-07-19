package bnet

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"log"
)

type PresenceServiceBinder struct{}

func (PresenceServiceBinder) Bind(sess *Session) Service {
	res := &PresenceService{}
	res.sess = sess
	return res
}

// The Presence service maintains and broadcasts the "presence" of players
// within the Bnet system: their current game, status, location, etc.
type PresenceService struct {
	sess *Session
}

func (s *PresenceService) Name() string {
	return "bnet.protocol.presence.PresenceService"
}

func (s *PresenceService) Methods() []string {
	return []string{
		"",
		"Subscribe",
		"Unsubscribe",
		"Update",
		"Query",
	}
}

func (s *PresenceService) Invoke(method int, body []byte) (resp []byte, err error) {
	switch method {
	case 1:
		return []byte{}, s.Subscribe(body)
	case 2:
		return []byte{}, s.Unsubscribe(body)
	case 3:
		return []byte{}, s.Update(body)
	case 4:
		return []byte{}, s.Query(body)
	default:
		log.Panicf("error: PresenceService.Invoke: unknown method %v", method)
		return
	}
}

func (s *PresenceService) Subscribe(body []byte) error {
	req := hsproto.BnetProtocolPresence_SubscribeRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())
	return nil
}

func (s *PresenceService) Unsubscribe(body []byte) error {
	return nyi
}

func (s *PresenceService) Update(body []byte) error {
	return nyi
}

func (s *PresenceService) Query(body []byte) error {
	return nyi
}
