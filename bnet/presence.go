package bnet

import (
	_ "github.com/HearthSim/hs-proto-go/bnet/attribute"
	"github.com/HearthSim/hs-proto-go/bnet/presence_service"
	"github.com/HearthSim/hs-proto-go/bnet/presence_types"
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
	req := presence_service.SubscribeRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())

	e := req.GetEntityId()
	e_key := amEntityId{e.GetHigh(), e.GetLow()}
	s.sess.server.accountManager.Subscribe(e_key, s.sess.account)
	log.Printf("Notification subscribe: %s subscribed to %s", EntityId(s.sess.account.EntityId.GetHigh(), s.sess.account.EntityId.GetLow()), e)

	return nil
}

func (s *PresenceService) Unsubscribe(body []byte) error {
	return nyi
}

func (s *PresenceService) Update(body []byte) error {
	req := presence_service.UpdateRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())

	e := req.GetEntityId()
	for _, fo := range req.GetFieldOperation() {
		f := fo.GetField()
		k := f.GetKey()
		//pk := PresenceKey{0,0,0,0,0,0}
		e_key := amEntityId{e.GetHigh(), e.GetLow()}
		p_key := amPresenceKey{k.GetProgram(), k.GetGroup(), k.GetField(), k.GetIndex()}
		switch fo.GetOperation() {
		case presence_types.FieldOperation_SET:
			// SET
			log.Printf("Presence field update [%s] = %s", fo, f.GetValue())
			//var v attribute.Variant
			v := *f.GetValue()
			s.sess.server.accountManager.UpdatePresenceData(e_key, p_key, v)
		case presence_types.FieldOperation_CLEAR:
			// assuming CLEAR
			// TODO: Test it
			log.Printf("Removing key from presence [%s]", fo)
			s.sess.server.accountManager.RemovePresenceData(e_key, p_key)
		default:
			// wrong case, should we raise error?
		}
	}

	return nil
}

func (s *PresenceService) Query(body []byte) error {
	req := presence_service.QueryRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())

	e := req.GetEntityId()
	e_key := *amEntityId_convert(e)
	res := presence_service.QueryResponse{}

	for _, k := range req.GetKey() {
		// build key
		p_key := amPresenceKey{k.GetProgram(), k.GetGroup(), k.GetField(), k.GetIndex()}
		log.Printf("Query for entity [%s %s]", e_key, p_key)
		v := s.sess.server.accountManager.GetPresenceData(e_key, p_key)
		log.Printf("Obtained value: %s", v)
		res.Field = append(res.Field, &presence_types.Field{
			Key:   k,
			Value: &v,
		})
	}

	// queue packet
	resBody, err := proto.Marshal(&res)
	if err != nil {
		return err
	}
	resHeader := s.sess.MakeRequestHeader(s, 1, len(resBody))
	err = s.sess.QueuePacket(resHeader, resBody)
	if err != nil {
		return err
	}

	return nyi
}
