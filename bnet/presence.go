package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/presence_service"
	"github.com/HearthSim/hs-proto-go/bnet/presence_types"
	"github.com/HearthSim/hs-proto-go/bnet/channel_types"
	"github.com/HearthSim/hs-proto-go/bnet/channel_service"
	_ "github.com/HearthSim/hs-proto-go/bnet/attribute"
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
	e_key := amEntityId{e.GetHigh(),e.GetLow()}
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
			p_key := amPresenceKey{e.GetHigh(), e.GetLow(), k.GetProgram(),k.GetGroup(),k.GetField(),k.GetIndex()}
			switch fo.GetOperation() {
				case presence_types.FieldOperation_SET:
					// SET
					log.Printf("Presence field update [%s] = %s", fo, f.GetValue())
					//var v attribute.Variant
					v := *f.GetValue()
					s.sess.server.accountManager.UpdatePresenceData(p_key, v)
					if k.GetGroup() == 2 && k.GetField() == 1 {
						// online status?
						log.Printf("Received online status: [%s]", v) 
					}
				case presence_types.FieldOperation_CLEAR:
					// assuming CLEAR
					// TODO: Test it
					log.Printf("Removing key from presence [%s]", fo)
					s.sess.server.accountManager.RemovePresenceData(p_key)
				default:
					// wrong case, should we raise error?
			}
	}

	// notify all subscribers
	e_key := amEntityId{e.GetHigh(), e.GetLow()}
	subs, ok := s.sess.server.accountManager.Subscribers[e_key]
	if !ok {
		log.Printf("No subscribers found.")
	} else {
		for _, sub := range subs {
			log.Printf("Notifying subscriber [%+v]", sub)
			log.Printf("Session state [%+v]", sub.Session.state)
			if sub.Session.state == 3 {
				//session is connected
				state := presence_types.ChannelState{}
				state.EntityId = e
				state.FieldOperation = req.GetFieldOperation()
				log.Printf("state: [%+v]", state)
				channel_state := channel_types.ChannelState{}
				//channel_state.ExtensionMap()

				proto.SetExtension(&channel_state, presence_types.E_ChannelState_Presence, &state)
				log.Printf("channel_state: [%+v] [%+v]", channel_state, channel_state.XXX_extensions)

				res := channel_service.UpdateChannelStateNotification{}
				res.StateChange = &channel_state
				log.Printf("res: [%s]", res)
				buf, err := proto.Marshal(&channel_state)
				if err != nil {
					panic(err)
				}
				log.Printf("buf: [%+v]", buf)
				/*
				header := s.sess.MakeRequestHeader(s, 1, len(buf))
				err = s.sess.QueuePacket(header, buf)
				if err != nil {
					panic(err)
				}*/
				s.sess.Respond(9999, buf)
			} else {
				log.Printf("Not notifying as session is probably not connected")
			}
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
	res := presence_service.QueryResponse{}

	for _, k := range req.GetKey() {
			// build key
			p_key := amPresenceKey{e.GetHigh(), e.GetLow(), k.GetProgram(),k.GetGroup(),k.GetField(),k.GetIndex()}
			log.Printf("Query for entity [%+v]", p_key)
			v := s.sess.server.accountManager.GetPresenceData(p_key)
			log.Printf("Obtained value: %s", v)
			res.Field = append(res.Field, &presence_types.Field{
					Key: k,
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

	return nil
}


