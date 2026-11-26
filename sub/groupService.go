package sub

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/service"
	"github.com/alireza0/s-ui/util"
	"github.com/alireza0/s-ui/util/common"
)

// GroupSubService serves both new token-protected routes. Group and personal
// links share the existing single-client builders, so formats and disable
// rules stay identical; only the set of clients and the authorization differ.
type GroupSubService struct {
	SubService
	JsonService
	ClashService
}

func (s *GroupSubService) groupClients(subID, token string) (*model.SubscriptionGroup, []*model.Client, error) {
	group, err := service.LoadSubscriptionGroupBySubID(nil, subID)
	if err != nil || group.Revoked || !service.CheckSubscriptionToken(group.TokenEnc, token) {
		return nil, nil, common.NewError("subscription not found")
	}
	clients, err := service.SubscriptionGroupClients(nil, group.Id)
	if err != nil {
		return nil, nil, err
	}
	return group, clients, nil
}

func (s *GroupSubService) memberClient(subID, token string) (*model.Client, error) {
	member, err := service.LoadSubscriptionMemberBySubID(nil, subID)
	if err != nil || member.Revoked || !service.CheckSubscriptionToken(member.TokenEnc, token) {
		return nil, common.NewError("subscription not found")
	}
	return s.SubService.getClientByID(member.ClientId)
}

// GroupContent returns one complete group response in the requested format.
func (s *GroupSubService) GroupContent(subID, token, format string) (*string, []string, error) {
	group, clients, err := s.groupClients(subID, token)
	if err != nil {
		return nil, nil, err
	}
	headers := s.groupHeaders(group, clients)
	switch format {
	case "json":
		inbounds := make([][]*model.Inbound, len(clients))
		for i, client := range clients {
			inbounds[i], err = GetClientInbounds(client)
			if err != nil {
				return nil, nil, err
			}
		}
		return s.JsonService.GetJsonForClients(clients, inbounds, headers)
	case "clash":
		inbounds := make([][]*model.Inbound, len(clients))
		for i, client := range clients {
			inbounds[i], err = GetClientInbounds(client)
			if err != nil {
				return nil, nil, err
			}
		}
		return s.ClashService.GetClashForClients(clients, inbounds, headers)
	case "happ":
		return s.groupLinks(clients, headers, happLinks)
	default:
		return s.groupLinks(clients, headers, nil)
	}
}

func (s *GroupSubService) groupLinks(clients []*model.Client, headers []string, transform func([]string) []string) (*string, []string, error) {
	var links []string
	for _, client := range clients {
		clientLinks, err := s.SubService.getClientLinks(client)
		if err != nil {
			return nil, nil, err
		}
		links = append(links, clientLinks...)
	}
	if transform != nil {
		links = transform(links)
	}
	result := strings.Join(links, "\n")
	subEncode, _ := s.SubService.SettingService.GetSubEncode()
	if subEncode {
		result = base64.StdEncoding.EncodeToString([]byte(result))
	}
	return &result, headers, nil
}

func (s *GroupSubService) groupHeaders(group *model.SubscriptionGroup, clients []*model.Client) []string {
	interval, _ := s.SubService.SettingService.GetSubUpdates()
	title := group.Name
	if strings.TrimSpace(title) == "" {
		title = "Subscription"
	}
	// Sum quotas/usage and use the earliest expiry only when every member has
	// both a finite quota and a finite expiry. Mixing limited and unlimited
	// members would make a single aggregate number lie, so the optional
	// Subscription-Userinfo header is omitted in that case.
	bounded := true
	var volume, up, down, expiry int64
	for i, client := range clients {
		if client.Volume <= 0 || client.Expiry <= 0 {
			bounded = false
			break
		}
		volume += client.Volume
		up += client.Up
		down += client.Down
		if i == 0 || client.Expiry < expiry {
			expiry = client.Expiry
		}
	}
	userinfo := ""
	if bounded && len(clients) > 0 {
		synthetic := &model.Client{Name: title, Remark: title, Volume: volume, Up: up, Down: down, Expiry: expiry}
		userinfo = util.GetHeaders(synthetic, interval)[0]
	}
	return []string{userinfo, strconv.Itoa(interval), title}
}

// PersonalContent returns only the member's own current configuration.
func (s *GroupSubService) PersonalContent(subID, token, format string) (*string, []string, error) {
	client, err := s.memberClient(subID, token)
	if err != nil {
		return nil, nil, err
	}
	inbounds, err := GetClientInbounds(client)
	if err != nil {
		return nil, nil, err
	}
	headers := s.SubService.getClientHeaders(client)
	switch format {
	case "json":
		return s.JsonService.GetJsonForClients([]*model.Client{client}, [][]*model.Inbound{inbounds}, headers)
	case "clash":
		return s.ClashService.GetClashForClients([]*model.Client{client}, [][]*model.Inbound{inbounds}, headers)
	case "happ":
		return s.SubService.getSubsForClient(client, happLinks)
	default:
		return s.SubService.getSubsForClient(client, nil)
	}
}

// GroupSummary is the admin view of one group. Links are only filled for an
// explicit authorised read.
type GroupSummary struct {
	Id             uint            `json:"id"`
	SubscriptionID string          `json:"subscriptionId"`
	Name           string          `json:"name"`
	BatchId        string          `json:"batchId"`
	Revoked        bool            `json:"revoked"`
	CreatedAt      int64           `json:"createdAt"`
	MemberCount    int             `json:"memberCount"`
	Link           string          `json:"link,omitempty"`
	Members        []MemberSummary `json:"members,omitempty"`
}

type MemberSummary struct {
	Id             uint   `json:"id"`
	SubscriptionID string `json:"subscriptionId"`
	ClientID       uint   `json:"clientId"`
	ClientName     string `json:"clientName"`
	Enabled        bool   `json:"enabled"`
	Revoked        bool   `json:"revoked"`
	Link           string `json:"link,omitempty"`
}

func (s *GroupSubService) GroupSummary(groupID uint, hostname string) (*GroupSummary, error) {
	group, err := service.LoadSubscriptionGroup(nil, groupID)
	if err != nil {
		return nil, err
	}
	sum := &GroupSummary{
		Id: group.Id, SubscriptionID: group.SubID, Name: group.Name, BatchId: group.BatchId,
		Revoked: group.Revoked, CreatedAt: group.CreatedAt,
	}
	base, _ := service.SubscriptionLinkBase(hostname)
	if link, err := service.GroupSubscriptionLink(base, group); err == nil {
		sum.Link = link
	}
	members, err := service.SubscriptionGroupMembers(nil, group.Id)
	if err != nil {
		return nil, err
	}
	sum.MemberCount = len(members)
	for _, member := range members {
		var client model.Client
		_ = database.GetDB().Select("id", "name", "enable").First(&client, member.ClientId).Error
		ms := MemberSummary{
			Id: member.Id, SubscriptionID: member.SubID, ClientID: member.ClientId,
			ClientName: client.Name, Enabled: client.Enable, Revoked: member.Revoked,
		}
		if link, err := service.MemberSubscriptionLink(base, member); err == nil {
			ms.Link = link
		}
		sum.Members = append(sum.Members, ms)
	}
	return sum, nil
}

// ListGroups returns group metadata without links or members. Authorised
// per-group reads use GroupSummary, which is the only place links appear.
func (s *GroupSubService) ListGroups() ([]GroupSummary, error) {
	var groups []model.SubscriptionGroup
	if err := database.GetDB().Order("id desc").Find(&groups).Error; err != nil {
		return nil, err
	}
	out := make([]GroupSummary, 0, len(groups))
	for _, group := range groups {
		var count int64
		_ = database.GetDB().Model(model.SubscriptionMember{}).Where("group_id = ?", group.Id).Count(&count).Error
		out = append(out, GroupSummary{
			Id: group.Id, SubscriptionID: group.SubID, Name: group.Name, BatchId: group.BatchId,
			Revoked: group.Revoked, CreatedAt: group.CreatedAt, MemberCount: int(count),
		})
	}
	return out, nil
}
