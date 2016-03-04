package bnet

import (
	"github.com/HearthSim/hs-proto-go/bnet/entity"
	"github.com/HearthSim/hs-proto-go/bnet/friends_service"
	"github.com/HearthSim/hs-proto-go/bnet/friends_types"
	"github.com/HearthSim/hs-proto-go/bnet/invitation_types"
	"github.com/HearthSim/hs-proto-go/bnet/role"
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

	// get user EntityId from session data
	log.Printf("FriendsService: SubscribeToFriends: Account EntityId [%s]", s.sess.account.ID)

	log.Printf("req = %s", req.String())

	res := friends_service.SubscribeToFriendsResponse{}
	res.MaxFriends = proto.Uint32(200)
	res.MaxReceivedInvitations = proto.Uint32(1000)
	res.MaxSentInvitations = proto.Uint32(20)

	// add roles
	res.Role = []*role.Role{
		&role.Role{Id: proto.Uint32(1), Name: proto.String("battle_tag_friend")},
		&role.Role{Id: proto.Uint32(2), Name: proto.String("real_id_friend")},
	}

	// handle friends
	friendIDs := struct{ ID uint64 }{}
	db.Table("friend").Select("Target as ID").Where("Source = ?", s.sess.account.ID).Scan(&friendIDs)
	friends := []Account{}
	db.Where(friendIDs).Find(&friends)

	for _, friend := range friends {
		// TODO: add real name, instead using battleTag?
		f := friends_types.Friend{
			Id:        EntityId(BnetAccountEntityIDHi, friend.ID),
			FullName:  proto.String(friend.BattleTag),
			BattleTag: proto.String(friend.BattleTag),
			Role:      []uint32{1},
		}
		res.Friends = append(res.Friends, &f)
	}

	// handle invitations
	res.ReceivedInvitations = []*invitation_types.Invitation{}
	invitationRequests := []InvitationRequest{}
	db.Table("invitation_request").Where("invitee_id = ?", s.sess.account.ID).Find(&invitationRequests)

	for _, ir := range invitationRequests {
		// TODO: Check for expired invitations
		accounts := []Account{}
		db.Where("id = ?", ir.InviterID).First(&accounts)
		if len(accounts) > 0 {
			res.ReceivedInvitations = append(
				res.ReceivedInvitations,
				&invitation_types.Invitation{
					Id:          proto.Uint64(ir.ID),
					InviterName: proto.String(accounts[0].BattleTag),
					InviterIdentity: &entity.Identity{
						AccountId: EntityId(BnetAccountEntityIDHi, accounts[0].ID),
					},
					InviteeIdentity: &entity.Identity{
						AccountId: EntityId(BnetAccountEntityIDHi, s.sess.account.ID),
					},
				})
		} else {
			// TODO: Remove invalid invitation from DB
			log.Printf("Account %+v not found. Skipping this invitation.", ir)
		}
	}

	res.SentInvitations = []*invitation_types.Invitation{}

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
	req := invitation_types.GenericRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())

	ir := []InvitationRequest{}
	db.Where("id = ?", req.GetInvitationId()).First(&ir)

	if len(ir) > 0 {
		friend1 := Friend{
			Source: ir[0].InviterID,
			Target: ir[0].InviteeID,
		}
		db.Create(&friend1)
		friend2 := Friend{
			Source: ir[0].InviteeID,
			Target: ir[0].InviterID,
		}
		db.Create(&friend2)
		db.Delete(&ir[0])
	} else {
		log.Printf("Invitation request not found")
	}

	// TODO: Notify both accounts (if online) about accepted friendship

	return nil
}

func (s *FriendsService) RevokeInvitation(body []byte) error {
	log.Printf("FriendsService: Revoke Invitation")
	return nyi
}

func (s *FriendsService) DeclineInvitation(body []byte) error {
	log.Printf("FriendsService: Decline Invitation")
	req := invitation_types.GenericRequest{}
	err := proto.Unmarshal(body, &req)
	if err != nil {
		return err
	}
	log.Printf("req = %s", req.String())

	ir := InvitationRequest{}
	db.Where("id = ?", req.GetInvitationId()).First(&ir)
	db.Delete(&ir)

	// TODO: Notify both accounts about declined invitation
	return nil
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
