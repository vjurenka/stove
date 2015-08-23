package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/attribute"
	"github.com/HearthSim/hs-proto-go/bnet/entity"
	"github.com/golang/protobuf/proto"
	"log"
)

const (
	NotifyClientRequest    = "GS_CL_REQ"
	NotifyClientResponse   = "GS_CL_RES"
	NotifyFindGameRequest  = "GS_FG_REQ"
	NotifyFindGameResponse = "GS_FG_RES"
	NotifyQueueEntry       = "GQ_ENTRY"
	NotifyQueueUpdate      = "GQ_UPDATE"
	NotifyQueueExit        = "GQ_EXIT"
	NotifyQueueResult      = "G_RESULT"
	NotifyMatchMakerStart  = "MM_START"
	NotifyMatchMakerEnd    = "MM_END"
	NotifyWhisper          = "WHISPER"
	NotifySpectatorInvite  = "WTCG.SpectatorInvite"
)

// A notification sent or received by battle.net from or to another server.
type Notification struct {
	Type       string
	Attributes []*attribute.Attribute
}

// Wrapper type for disambiguationg blob and message values in attributes.
type MessageValue struct {
	Value []byte
}

// Wrapper type for disambiguationg string and fourcc values in attributes.
type FourccValue struct {
	Value string
}

func NewNotification(ty string, m map[string]interface{}) *Notification {
	res := &Notification{}
	res.Type = ty
	for k, v := range m {
		variant := &attribute.Variant{}
		switch v := v.(type) {
		case bool:
			variant.BoolValue = proto.Bool(v)
		// Minor annoyance here to have to do these casts, but it would be
		// a huge annoyance elsewhere:
		case int:
			variant.IntValue = proto.Int64(int64(v))
		case int32:
			variant.IntValue = proto.Int64(int64(v))
		case int64:
			variant.IntValue = proto.Int64(v)
		case uint:
			variant.UintValue = proto.Uint64(uint64(v))
		case uint32:
			variant.UintValue = proto.Uint64(uint64(v))
		case uint64:
			variant.UintValue = proto.Uint64(v)
		case float32:
			variant.FloatValue = proto.Float64(float64(v))
		case float64:
			variant.FloatValue = proto.Float64(v)
		case string:
			variant.StringValue = proto.String(v)
		case []byte:
			variant.BlobValue = v
		case MessageValue:
			variant.MessageValue = v.Value
		case FourccValue:
			variant.FourccValue = proto.String(v.Value)
		case entity.EntityId:
			variant.EntityidValue = &v
		default:
			log.Panicf("error: can't convert %s: %T to attribute", k, v)
		}
		res.Attributes = append(res.Attributes, &attribute.Attribute{
			Name:  proto.String(k),
			Value: variant,
		})
	}
	return res
}

// Converts a notification into a flat map containing a type key for the
// notification's type and all of the notification's attributes.
func (n *Notification) Map() map[string]interface{} {
	res := map[string]interface{}{}
	res["type"] = n.Type
	for _, attr := range n.Attributes {
		k := *attr.Name
		v := *attr.Value
		switch {
		case v.BoolValue != nil:
			res[k] = *v.BoolValue
		case v.IntValue != nil:
			res[k] = *v.IntValue
		case v.UintValue != nil:
			res[k] = *v.UintValue
		case v.FloatValue != nil:
			res[k] = *v.FloatValue
		case v.StringValue != nil:
			res[k] = *v.StringValue
		case v.BlobValue != nil:
			res[k] = v.BlobValue
		case v.MessageValue != nil:
			res[k] = MessageValue{v.MessageValue}
		case v.FourccValue != nil:
			res[k] = FourccValue{*v.FourccValue}
		case v.EntityidValue != nil:
			res[k] = *v.EntityidValue
		default:
			log.Panicf("error: variant(%s) has no value", k)
		}
	}
	return res
}

func (s *Session) HandleNotifications() {
	quit := s.ChanForTransition(StateDisconnected)
	for {
		select {
		case notify := <-s.ClientNotifications:
			s.handleNotification(notify)
		case <-quit:
			return
		}
	}
}

func (s *Session) handleNotification(notify *Notification) {
	log.Printf("received notification (%s): %v\n", notify.Type, notify.Attributes)
	switch notify.Type {
	default:
		if ch, ok := notificationHandlers[notify.Type]; ok {
			(<-ch)(notify)
			return
		}
		log.Panicf("error: unhandled notification type %s", notify.Type)
	}
}

var notificationHandlers = map[string]chan func(*Notification){}

// Will trigger handler once the server is notified with a notification of type
// ty.
func (s *Session) OnceNotified(ty string, handle func(n *Notification)) {
	if ch, ok := notificationHandlers[ty]; ok {
		ch <- handle
	} else {
		ch = make(chan func(*Notification), 1)
		notificationHandlers[ty] = ch
		ch <- handle
	}
}
