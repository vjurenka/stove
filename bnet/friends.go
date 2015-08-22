package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/friends_service"
	"github.com/golang/protobuf/proto"
	"log"
)

type FriendsServiceBinder struct{}

func (FriendsServiceBinder) Bind(sess *Session) Service {
	res := &FriendsService{}
	res.sess = sess
	return res
}

// The Friends service handles friends and friend requests.
type FriendsService struct {
	sess *Session
}

func (s *FriendsService) Name() string {
	return "bnet.protocol.friends.FriendsService"
}

func (s *FriendsService) Methods() []string {
	return []string{
		"",
		"SubscribeToFriends",
		"SendInvitation",
		"AcceptInvitation",
		"RevokeInvitation",
		"DeclineInvitation",
		"IgnoreInvitation",
		"AssignRole",
		"RemoveFriend",
		"ViewFriends",
		"UpdateFriendState",
		"UnsubscribeToFriends",
		"RevokeAllInvitations",
	}
}

func (s *FriendsService) Invoke(method int, body []byte) (resp []byte, err error) {
	switch method {
	case 1:
		return s.SubscribeToFriends(body)
	case 2:
		return []byte{}, s.SendInvitation(body)
	case 3:
		return []byte{}, s.AcceptInvitation(body)
	case 4:
		return []byte{}, s.RevokeInvitation(body)
	case 5:
		return []byte{}, s.DeclineInvitation(body)
	case 6:
		return []byte{}, s.IgnoreInvitation(body)
	case 7:
		return []byte{}, s.AssignRole(body)
	case 8:
		return s.RemoveFriend(body)
	case 9:
		return s.ViewFriends(body)
	case 10:
		return []byte{}, s.UpdateFriendState(body)
	case 11:
		return []byte{}, s.UnsubscribeToFriends(body)
	case 12:
		return []byte{}, s.RevokeAllInvitations(body)
	default:
		log.Panicf("error: FriendsService.Invoke: unknown method %v", method)
		return
	}
}

func (s *FriendsService) SubscribeToFriends(body []byte) ([]byte, error) {
	req := friends_service.SubscribeToFriendsRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	log.Printf("req = %s", req.String())
	res := friends_service.SubscribeToFriendsResponse{}
	res.MaxFriends = proto.Uint32(200)
	res.MaxReceivedInvitations = proto.Uint32(1000)
	res.MaxSentInvitations = proto.Uint32(20)
	resBuf, err := proto.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return resBuf, nil
}

func (s *FriendsService) SendInvitation(body []byte) error {
	return nyi
}

func (s *FriendsService) AcceptInvitation(body []byte) error {
	return nyi
}

func (s *FriendsService) RevokeInvitation(body []byte) error {
	return nyi
}

func (s *FriendsService) DeclineInvitation(body []byte) error {
	return nyi
}

func (s *FriendsService) IgnoreInvitation(body []byte) error {
	return nyi
}

func (s *FriendsService) AssignRole(body []byte) error {
	return nyi
}

func (s *FriendsService) RemoveFriend(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *FriendsService) ViewFriends(body []byte) ([]byte, error) {
	return nil, nyi
}

func (s *FriendsService) UpdateFriendState(body []byte) error {
	return nyi
}

func (s *FriendsService) UnsubscribeToFriends(body []byte) error {
	return nyi
}

func (s *FriendsService) RevokeAllInvitations(body []byte) error {
	return nyi
}
