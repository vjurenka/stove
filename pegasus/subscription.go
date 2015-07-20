package pegasus

import (
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"time"
)

// The subscription system is a way to reduce the server resource usage of
// idling players.
type Subscription struct {
	subscribed time.Time
	timeout    time.Duration
	route      uint64
}

func (s *Subscription) Init(sess *Session) {
	sess.RegisterUtilHandler(0, 314, OnUtilSubscribe)
}

func (s *Subscription) IsValid() bool {
	if s.timeout == 0 {
		return true
	}
	return time.Now().Before(s.subscribed.Add(s.timeout + 15*time.Second))
}

// Handle a client's subscribe request.  The response may specify a timeout,
// after which an active client must resubscribe to renew their session.
func OnUtilSubscribe(s *Session, body []byte) ([]byte, error) {
	if s.timeout == 0 {
		s.timeout = 120 * time.Second
	}
	if s.route == 0 {
		s.route = 1
	}
	s.subscribed = time.Now()
	res := hsproto.PegasusUtil_SubscribeResponse{}
	res.Route = proto.Uint64(s.route)
	res.SupportedFeatures = proto.Uint64(3)
	res.KeepAliveSecs = proto.Uint64(uint64(s.timeout.Seconds()))
	resBody, err := proto.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return EncodeUtilResponse(315, resBody)
}
