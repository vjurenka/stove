package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/friends_service"
	"github.com/HearthSim/hs-proto-go/bnet/friends_types"
	"github.com/HearthSim/hs-proto-go/bnet/role"
	"github.com/HearthSim/hs-proto-go/bnet/entity"
	"github.com/HearthSim/hs-proto-go/bnet/invitation_types"
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

	e := s.sess.account.EntityId
	//e_high := 72058118023938048
	log.Printf("FriendsService: SubscribeToFriends: Account EntityId [%s]", e)

	log.Printf("req = %s", req.String())
	res := friends_service.SubscribeToFriendsResponse{}
	res.MaxFriends = proto.Uint32(200)
	res.MaxReceivedInvitations = proto.Uint32(1000)
	res.MaxSentInvitations = proto.Uint32(20)

	roles := []*role.Role{}
	role1 := role.Role{}
	role1.Id = proto.Uint32(1)
	role1.Name = proto.String("battle_tag_friend")
	role2 := role.Role{}
	role2.Id = proto.Uint32(2)
	role2.Name = proto.String("real_id_friend")

	roles = append(roles, &role1)
	roles = append(roles, &role2)
	res.Role = roles

	// TODO: collect friends from DB
	friends := []*friends_types.Friend{}
	// example friend :)
	friend1 := friends_types.Friend{}
	friend1.Id = EntityId(72058118023938048, 1)
	friend1.FullName = proto.String("")
	friend1.BattleTag = proto.String("Test#1234")
	friend2 := friends_types.Friend{}
	friend2.Id = EntityId(72058118023938048, 2)
	//friend2.FullName = proto.String("")
	friend2.BattleTag = proto.String("Quoing#1000")

	friends = append(friends, &friend1)
	friends = append(friends, &friend2)
	res.Friends = friends

	sentInvitations := []*invitation_types.Invitation{}
	receivedInvitations := []*invitation_types.Invitation{}


	invitationRequests := []InvitationRequest{}
	db.Where("invitee_id = ?", e.GetLow()).Find(&invitationRequests)
	inv_id := uint64(1)

	for _,invs := range invitationRequests {
		log.Printf("Found invitation request [%+v]", invs)
		accounts := []Account{}
		db.Where("id = ?", invs.ID).First(&accounts)
		log.Printf("...invitation from [%+v]", accounts)
		ri := invitation_types.Invitation{}
		ri.Id = proto.Uint64(inv_id)
		inv_id += 1
		ri.InviterName = proto.String(accounts[0].BattleTag)
		ri.InviterIdentity = &entity.Identity{AccountId: EntityId(72058118023938048, accounts[0].ID),}
		ri.InviteeIdentity = &entity.Identity{AccountId: EntityId(72058118023938048, e.GetLow()),}
		log.Printf("Adding invitation [%+v]", ri)
		receivedInvitations = append(receivedInvitations, &ri)
	}

	res.SentInvitations = sentInvitations
	res.ReceivedInvitations = receivedInvitations

	resBuf, err := proto.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return resBuf, nil
}

func (s *FriendsService) SendInvitation(body []byte) error {
	log.Printf("FriendService: Send Invitation")
	return nyi
}

func (s *FriendsService) AcceptInvitation(body []byte) error {
	log.Printf("FriendsService: Accept Invitation")
	log.Printf("req = %+v", body)
	req := invitation_types.GenericRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
			return err
	}
	log.Printf("req = %s", req.String())

	iReq := InvitationRequest{}

	db.Where("id = ?", req.GetInvitationId()).First(&iReq)

	friend := Friend{}
	friend.ID = iReq.InviterID
	friend.FriendID = iReq.InviteeID
	db.Create(friend)
	friend.ID = iReq.InviteeID
	friend.FriendID = iReq.InviterID
	db.Create(friend)

	db.Delete(&iReq)

	// TODO: Notify both accounts (if online) about accepted friendship

	return nil
}

func (s *FriendsService) RevokeInvitation(body []byte) error {
	log.Printf("FriendsService: Revoke Invitation")
	return nyi
}

func (s *FriendsService) DeclineInvitation(body []byte) error {
	log.Printf("FriendsService: Decline Invitation")
	return nyi
}

func (s *FriendsService) IgnoreInvitation(body []byte) error {
	log.Printf("FriendsService: Ignore Invitation")
	return nyi
}

func (s *FriendsService) AssignRole(body []byte) error {
	log.Printf("FriendsService: Assign Role")
	return nyi
}

func (s *FriendsService) RemoveFriend(body []byte) ([]byte, error) {
	log.Printf("FriendsService: Remove Friend")
	return nil, nyi
}

func (s *FriendsService) ViewFriends(body []byte) ([]byte, error) {
	log.Printf("FriendsService: View Friends")
	return nil, nyi
}

func (s *FriendsService) UpdateFriendState(body []byte) error {
	log.Printf("FriendsService: Update friend state")
	return nyi
}

func (s *FriendsService) UnsubscribeToFriends(body []byte) error {
	log.Printf("FriendsService: Unsubscribe to friends")
	return nyi
}

func (s *FriendsService) RevokeAllInvitations(body []byte) error {
	log.Printf("FriendsService: Revoke All Invitations")
	return nyi
}
