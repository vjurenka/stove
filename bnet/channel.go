package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/channel_invitation_service"
	"github.com/golang/protobuf/proto"
	"log"
)

type ChannelInvitationServiceBinder struct{}

func (ChannelInvitationServiceBinder) Bind(sess *Session) Service {
	res := &ChannelInvitationService{}
	res.sess = sess
	return res
}

// The ChannelInvitation service manages chat channel invitations.
type ChannelInvitationService struct {
	sess *Session
}

func (s *ChannelInvitationService) Name() string {
	return "bnet.protocol.channel_invitation.ChannelInvitationService"
}

func (s *ChannelInvitationService) Methods() []string {
	return []string{
		"",
		"Subscribe",
		"Unsubscribe",
		"SendInvitation",
		"AcceptInvitation",
		"DeclineInvitation",
		"RevokeInvitation",
		"SuggestInvitation",
		"IncrementChannelCount",
		"DecrementChannelCount",
		"UpdateChannelCount",
		"ListChannelCount",
	}
}

func (s *ChannelInvitationService) Invoke(method int, body []byte) (resp []byte, err error) {
	switch method {
	case 1:
		return s.Subscribe(body)
	case 2:
		return []byte{}, s.Unsubscribe(body)
	case 3:
		return s.SendInvitation(body)
	case 4:
		return s.AcceptInvitation(body)
	case 5:
		return []byte{}, s.DeclineInvitation(body)
	case 6:
		return []byte{}, s.RevokeInvitation(body)
	case 7:
		return []byte{}, s.SuggestInvitation(body)
	case 8:
		return s.IncrementChannelCount(body)
	case 9:
		return []byte{}, s.DecrementChannelCount(body)
	case 10:
		return []byte{}, s.UpdateChannelCount(body)
	case 11:
		return s.ListChannelCount(body)
	default:
		log.Panicf("error: ChannelInvitationService.Invoke: unknown method %v", method)
		return
	}
}

func (s *ChannelInvitationService) Subscribe(body []byte) ([]byte, error) {
	req := channel_invitation_service.SubscribeRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	res := channel_invitation_service.SubscribeResponse{}
	resBuf, err := proto.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return resBuf, nil
}

func (s *ChannelInvitationService) Unsubscribe(body []byte) error {
	return nyi
}

func (s *ChannelInvitationService) SendInvitation(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *ChannelInvitationService) AcceptInvitation(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *ChannelInvitationService) DeclineInvitation(body []byte) error {
	return nyi
}

func (s *ChannelInvitationService) RevokeInvitation(body []byte) error {
	return nyi
}

func (s *ChannelInvitationService) SuggestInvitation(body []byte) error {
	return nyi
}

func (s *ChannelInvitationService) IncrementChannelCount(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *ChannelInvitationService) DecrementChannelCount(body []byte) error {
	return nyi
}

func (s *ChannelInvitationService) UpdateChannelCount(body []byte) error {
	return nyi
}

func (s *ChannelInvitationService) ListChannelCount(body []byte) ([]byte, error) {
	return nil, nyi
}
