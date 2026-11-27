package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/util/common"

	"gorm.io/gorm"
)

// subscriptionCipher derives a stable AES-GCM key from the panel secret. The
// secret lives in the settings table and travels with a database backup, so a
// restored installation can still decrypt the group/member tokens it copied.
// Tokens are never stored in the clear, and they are not hashes: an authorised
// administrator has to be able to copy the same link again later.
func subscriptionCipher() (cipher.AEAD, error) {
	secret, err := (&SettingService{}).GetSecret()
	if err != nil {
		return nil, err
	}
	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func encryptSubscriptionTokenWith(aead cipher.AEAD, token string) (string, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := aead.Seal(nil, nonce, []byte(token), nil)
	return base64.RawURLEncoding.EncodeToString(append(nonce, sealed...)), nil
}

func encryptSubscriptionToken(token string) (string, error) {
	aead, err := subscriptionCipher()
	if err != nil {
		return "", err
	}
	return encryptSubscriptionTokenWith(aead, token)
}

// DecryptSubscriptionToken returns the copyable token for an authorised admin.
func DecryptSubscriptionToken(encoded string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	aead, err := subscriptionCipher()
	if err != nil {
		return "", err
	}
	if len(raw) < aead.NonceSize() {
		return "", common.NewError("subscription token is truncated")
	}
	plain, err := aead.Open(nil, raw[:aead.NonceSize()], raw[aead.NonceSize():], nil)
	if err != nil {
		return "", common.NewError("subscription token cannot be decrypted; was the panel secret changed?")
	}
	return string(plain), nil
}

// CheckSubscriptionToken compares a supplied token in constant time.
func CheckSubscriptionToken(encoded, supplied string) bool {
	if supplied == "" {
		return false
	}
	expected, err := DecryptSubscriptionToken(encoded)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) == 1
}

func newSubscriptionID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func newSubscriptionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// CreateImportSubscriptions creates the group identity and one personal
// identity per imported client in the import transaction. It is called once
// per batch; replay loads the stored IDs instead of creating new rows.
func CreateImportSubscriptions(tx *gorm.DB, aead cipher.AEAD, batchID, name string, clients []*model.Client) (*model.SubscriptionGroup, map[uint]*model.SubscriptionMember, error) {
	groupID, err := newSubscriptionID()
	if err != nil {
		return nil, nil, err
	}
	token, err := newSubscriptionToken()
	if err != nil {
		return nil, nil, err
	}
	enc, err := encryptSubscriptionTokenWith(aead, token)
	if err != nil {
		return nil, nil, err
	}
	group := &model.SubscriptionGroup{
		SubID: groupID, Name: name, BatchId: batchID, TokenEnc: enc, CreatedAt: nowUnix(),
	}
	if err := tx.Create(group).Error; err != nil {
		return nil, nil, err
	}
	members := make(map[uint]*model.SubscriptionMember, len(clients))
	for _, client := range clients {
		memberID, err := newSubscriptionID()
		if err != nil {
			return nil, nil, err
		}
		memberToken, err := newSubscriptionToken()
		if err != nil {
			return nil, nil, err
		}
		memberEnc, err := encryptSubscriptionTokenWith(aead, memberToken)
		if err != nil {
			return nil, nil, err
		}
		member := &model.SubscriptionMember{
			GroupId: group.Id, ClientId: client.Id, SubID: memberID, TokenEnc: memberEnc, CreatedAt: nowUnix(),
		}
		if err := tx.Create(member).Error; err != nil {
			return nil, nil, err
		}
		members[client.Id] = member
	}
	return group, members, nil
}

func LoadSubscriptionGroup(db *gorm.DB, id uint) (*model.SubscriptionGroup, error) {
	if db == nil {
		db = database.GetDB()
	}
	var group model.SubscriptionGroup
	err := db.Where("id = ?", id).First(&group).Error
	return &group, err
}

func LoadSubscriptionGroupBySubID(db *gorm.DB, subID string) (*model.SubscriptionGroup, error) {
	if db == nil {
		db = database.GetDB()
	}
	var group model.SubscriptionGroup
	err := db.Where("sub_id = ?", subID).First(&group).Error
	return &group, err
}

func LoadSubscriptionMemberBySubID(db *gorm.DB, subID string) (*model.SubscriptionMember, error) {
	if db == nil {
		db = database.GetDB()
	}
	var member model.SubscriptionMember
	err := db.Where("sub_id = ?", subID).First(&member).Error
	return &member, err
}

func LoadSubscriptionMemberByClientID(db *gorm.DB, clientID uint) (*model.SubscriptionMember, error) {
	if db == nil {
		db = database.GetDB()
	}
	var member model.SubscriptionMember
	err := db.Where("client_id = ?", clientID).First(&member).Error
	return &member, err
}

// SubscriptionGroupMembers returns members in import order (member id).
func SubscriptionGroupMembers(db *gorm.DB, groupID uint) ([]*model.SubscriptionMember, error) {
	if db == nil {
		db = database.GetDB()
	}
	var members []*model.SubscriptionMember
	err := db.Where("group_id = ?", groupID).Order("id").Find(&members).Error
	return members, err
}

// SubscriptionGroupClients loads the currently subscribe-able clients of a
// group, preserving member/import order and applying the same enable check as
// the legacy subscription path. Deleted or disabled clients simply drop out.
func SubscriptionGroupClients(db *gorm.DB, groupID uint) ([]*model.Client, error) {
	if db == nil {
		db = database.GetDB()
	}
	members, err := SubscriptionGroupMembers(db, groupID)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint]*model.Client, len(members))
	ids := make([]uint, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.ClientId)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var clients []*model.Client
	if err := db.Model(model.Client{}).Where("id IN ? AND enable = ?", ids, true).Find(&clients).Error; err != nil {
		return nil, err
	}
	for _, client := range clients {
		byID[client.Id] = client
	}
	ordered := make([]*model.Client, 0, len(clients))
	for _, id := range ids {
		if client, ok := byID[id]; ok {
			ordered = append(ordered, client)
		}
	}
	return ordered, nil
}

func rotateSubscriptionToken(db *gorm.DB, table string, id uint) (string, error) {
	token, err := newSubscriptionToken()
	if err != nil {
		return "", err
	}
	enc, err := encryptSubscriptionToken(token)
	if err != nil {
		return "", err
	}
	if err := db.Table(table).Where("id = ?", id).Update("token_enc", enc).Error; err != nil {
		return "", err
	}
	return token, nil
}

func RotateGroupSubscriptionToken(groupID uint) (string, error) {
	return rotateSubscriptionToken(database.GetDB(), "subscription_groups", groupID)
}

func RotateMemberSubscriptionToken(memberID uint) (string, error) {
	return rotateSubscriptionToken(database.GetDB(), "subscription_members", memberID)
}

func SetGroupSubscriptionRevoked(groupID uint, revoked bool) error {
	return database.GetDB().Model(&model.SubscriptionGroup{}).Where("id = ?", groupID).Update("revoked", revoked).Error
}

func SetMemberSubscriptionRevoked(memberID uint, revoked bool) error {
	return database.GetDB().Model(&model.SubscriptionMember{}).Where("id = ?", memberID).Update("revoked", revoked).Error
}

func RenameSubscriptionGroup(groupID uint, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return common.NewError("group name must not be empty")
	}
	return database.GetDB().Model(&model.SubscriptionGroup{}).Where("id = ?", groupID).Update("name", name).Error
}

func RemoveSubscriptionMember(memberID uint) error {
	res := database.GetDB().Where("id = ?", memberID).Delete(&model.SubscriptionMember{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.NewError("no such subscription member")
	}
	return nil
}

// EnsureClientSubscription gives an older client its own personal identity on
// demand. A client without a group gets a single-member group named after it.
func EnsureClientSubscription(clientID uint) (*model.SubscriptionMember, *model.SubscriptionGroup, error) {
	db := database.GetDB()
	if member, err := LoadSubscriptionMemberByClientID(db, clientID); err == nil {
		group, gerr := LoadSubscriptionGroup(db, member.GroupId)
		return member, group, gerr
	} else if !database.IsNotFound(err) {
		return nil, nil, err
	}
	var client model.Client
	if err := db.Select("id", "name").First(&client, clientID).Error; err != nil {
		return nil, nil, err
	}
	aead, aeadErr := subscriptionCipher()
	if aeadErr != nil {
		return nil, nil, aeadErr
	}
	var group *model.SubscriptionGroup
	var member *model.SubscriptionMember
	err := db.Transaction(func(tx *gorm.DB) error {
		groupID, err := newSubscriptionID()
		if err != nil {
			return err
		}
		token, err := newSubscriptionToken()
		if err != nil {
			return err
		}
		enc, err := encryptSubscriptionTokenWith(aead, token)
		if err != nil {
			return err
		}
		group = &model.SubscriptionGroup{SubID: groupID, Name: client.Name, BatchId: "", TokenEnc: enc, CreatedAt: nowUnix()}
		if err := tx.Create(group).Error; err != nil {
			return err
		}
		memberID, err := newSubscriptionID()
		if err != nil {
			return err
		}
		memberToken, err := newSubscriptionToken()
		if err != nil {
			return err
		}
		memberEnc, err := encryptSubscriptionTokenWith(aead, memberToken)
		if err != nil {
			return err
		}
		member = &model.SubscriptionMember{GroupId: group.Id, ClientId: client.Id, SubID: memberID, TokenEnc: memberEnc, CreatedAt: nowUnix()}
		return tx.Create(member).Error
	})
	return member, group, err
}

// SubscriptionLinkBase normalises the configured subscription base address.
func SubscriptionLinkBase(host string) (string, error) {
	base, err := (&SettingService{}).GetFinalSubURI(host)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(base, "/") + "/", nil
}

func GroupSubscriptionLink(base string, group *model.SubscriptionGroup) (string, error) {
	if group == nil || group.Revoked {
		return "", nil
	}
	token, err := DecryptSubscriptionToken(group.TokenEnc)
	if err != nil {
		return "", err
	}
	return base + "g/" + group.SubID + "?token=" + token, nil
}

func MemberSubscriptionLink(base string, member *model.SubscriptionMember) (string, error) {
	if member == nil || member.Revoked {
		return "", nil
	}
	token, err := DecryptSubscriptionToken(member.TokenEnc)
	if err != nil {
		return "", err
	}
	return base + "c/" + member.SubID + "?token=" + token, nil
}

// EnrichImportResult adds subscription IDs and (when includeLinks is set) the
// copyable links to an import result. The stored batch result only carries the
// IDs, so a list/history read cannot leak a token.
func (s *ImportService) EnrichImportResult(db *gorm.DB, res *ImportResult, hostname string, includeLinks bool) {
	if res == nil || res.GroupSubscription.SubscriptionID == "" {
		return
	}
	group, err := LoadSubscriptionGroupBySubID(db, res.GroupSubscription.SubscriptionID)
	if err != nil {
		logger.Warning("unable to load subscription group: ", err)
		return
	}
	res.GroupSubscription.Name = group.Name
	res.GroupSubscription.Revoked = group.Revoked
	if includeLinks {
		base, err := SubscriptionLinkBase(hostname)
		if err != nil {
			logger.Warning("unable to build subscription base: ", err)
		} else if link, err := GroupSubscriptionLink(base, group); err == nil {
			res.GroupSubscription.Link = link
		}
	}
	base := ""
	if includeLinks {
		base, _ = SubscriptionLinkBase(hostname)
	}
	for i := range res.Items {
		if res.Items[i].SubscriptionID == "" {
			continue
		}
		var member model.SubscriptionMember
		if err := db.Where("sub_id = ?", res.Items[i].SubscriptionID).First(&member).Error; err != nil {
			continue
		}
		res.Items[i].SubscriptionRevoked = member.Revoked
		if includeLinks && base != "" {
			if link, err := MemberSubscriptionLink(base, &member); err == nil {
				res.Items[i].SubscriptionLink = link
			}
		}
	}
}

// DeleteClientSubscriptions removes the personal subscription row of every
// deleted client. A group that no longer has any members is removed too; a
// group shared with surviving clients keeps its identity and token.
func DeleteClientSubscriptions(tx *gorm.DB, clientIDs []uint) error {
	if len(clientIDs) == 0 {
		return nil
	}
	const chunk = 400
	groupIDs := make(map[uint]bool)
	for start := 0; start < len(clientIDs); start += chunk {
		end := start + chunk
		if end > len(clientIDs) {
			end = len(clientIDs)
		}
		var members []model.SubscriptionMember
		if err := tx.Where("client_id in ?", clientIDs[start:end]).Find(&members).Error; err != nil {
			return err
		}
		for _, member := range members {
			if member.GroupId != 0 {
				groupIDs[member.GroupId] = true
			}
		}
		if err := tx.Where("client_id in ?", clientIDs[start:end]).
			Delete(&model.SubscriptionMember{}).Error; err != nil {
			return err
		}
	}
	ids := make([]uint, 0, len(groupIDs))
	for id := range groupIDs {
		ids = append(ids, id)
	}
	for start := 0; start < len(ids); start += chunk {
		end := start + chunk
		if end > len(ids) {
			end = len(ids)
		}
		if err := tx.Where("id in ?", ids[start:end]).
			Where("NOT EXISTS (SELECT 1 FROM subscription_members WHERE subscription_members.group_id = subscription_groups.id)").
			Delete(&model.SubscriptionGroup{}).Error; err != nil {
			return err
		}
	}
	return nil
}
