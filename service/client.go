package service

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/alireza0/s-ui/core"
	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/util"
	"github.com/alireza0/s-ui/util/common"
	"github.com/alireza0/s-ui/util/rateconv"

	"gorm.io/gorm"
)

type ClientService struct{}

func (s *ClientService) Get(id string) (*[]model.Client, error) {
	if id == "" {
		return s.GetAll()
	}
	return s.getById(id)
}

func (s *ClientService) getById(id string) (*[]model.Client, error) {
	db := database.GetDB()
	var client []model.Client
	err := db.Model(model.Client{}).Where("id in ?", strings.Split(id, ",")).Scan(&client).Error
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (s *ClientService) GetAll() (*[]model.Client, error) {
	db := database.GetDB()
	var clients []model.Client
	err := db.Model(model.Client{}).
		Select("`id`, `enable`, `name`, `desc`, `group`, `remark`, `inbounds`, `up`, `down`, `volume`, `expiry`, `created_at`, `online_at`, `max_conns`, `ip_limit`, `up_mbps`, `down_mbps`").
		Scan(&clients).Error
	if err != nil {
		return nil, err
	}
	return &clients, nil
}

// GetMaxConnsMap returns a map of client name -> max concurrent connections
// for all clients that have a positive connection limit configured.
func (s *ClientService) GetMaxConnsMap(db *gorm.DB) (map[string]int, error) {
	if db == nil {
		db = database.GetDB()
	}
	var rows []struct {
		Name     string
		MaxConns int
	}
	err := db.Model(model.Client{}).
		Select("`name`, `max_conns`").
		Where("`max_conns` > 0").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	limits := make(map[string]int, len(rows))
	for _, row := range rows {
		limits[row.Name] = row.MaxConns
	}
	return limits, nil
}

// GetIPLimitMap returns the distinct-IP limits by client name for all clients
// that have a positive limit.
func (s *ClientService) GetIPLimitMap(db *gorm.DB) (map[string]int, error) {
	if db == nil {
		db = database.GetDB()
	}
	var rows []struct {
		Name    string
		IPLimit int
	}
	err := db.Model(model.Client{}).
		Select("`name`, `ip_limit`").
		Where("`ip_limit` > 0").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	limits := make(map[string]int, len(rows))
	for _, row := range rows {
		limits[row.Name] = row.IPLimit
	}
	return limits, nil
}

func clientIPLimitPresent(raw json.RawMessage) bool {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return true
	}
	_, present := fields["ipLimit"]
	return present
}

// preserveMissingIPLimit keeps the stored limit when a caller that predates
// the field omits it. Sending an explicit 0 clears it.
func (s *ClientService) preserveMissingIPLimit(tx *gorm.DB, client *model.Client, raw json.RawMessage) {
	if clientIPLimitPresent(raw) {
		return
	}
	var existing model.Client
	if err := tx.Select("id", "ip_limit").First(&existing, client.Id).Error; err == nil {
		client.IPLimit = existing.IPLimit
	}
}

// GetRateLimitsMap returns the non-zero bandwidth caps by client name. The
// query does not look at max_conns, so a client that only sets a rate is
// still loaded into the tracker.
func (s *ClientService) GetRateLimitsMap(db *gorm.DB) (map[string]core.UserRate, error) {
	if db == nil {
		db = database.GetDB()
	}
	var rows []struct {
		Name     string
		UpMbps   int
		DownMbps int
	}
	err := db.Model(model.Client{}).
		Select("`name`, `up_mbps`, `down_mbps`").
		Where("`up_mbps` > 0 OR `down_mbps` > 0").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	rates := make(map[string]core.UserRate, len(rows))
	for _, row := range rows {
		if int64(row.UpMbps) > rateconv.MaxMbps || int64(row.DownMbps) > rateconv.MaxMbps {
			return nil, common.NewErrorf("client %q rate exceeds the maximum of %d Mbps", row.Name, rateconv.MaxMbps)
		}
		rates[row.Name] = core.UserRate{UpMbps: row.UpMbps, DownMbps: row.DownMbps}
	}
	return rates, nil
}

func validateClientRates(client *model.Client) error {
	if client.UpMbps < 0 {
		client.UpMbps = 0
	}
	if client.DownMbps < 0 {
		client.DownMbps = 0
	}
	if int64(client.UpMbps) > rateconv.MaxMbps {
		return common.NewErrorf("upMbps must not exceed %d", rateconv.MaxMbps)
	}
	if int64(client.DownMbps) > rateconv.MaxMbps {
		return common.NewErrorf("downMbps must not exceed %d", rateconv.MaxMbps)
	}
	return nil
}

// clientRateFieldsPresent reports whether the request carried each rate field.
// An absent field must keep the stored value; an explicit 0 must clear it.
func clientRateFieldsPresent(raw json.RawMessage) (up, down bool) {
	if len(raw) == 0 {
		return true, true
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return true, true
	}
	_, up = fields["upMbps"]
	_, down = fields["downMbps"]
	return
}

func (s *ClientService) preserveMissingRates(tx *gorm.DB, client *model.Client, raw json.RawMessage) {
	up, down := clientRateFieldsPresent(raw)
	if up && down {
		return
	}
	var existing model.Client
	if err := tx.Select("id", "up_mbps", "down_mbps").First(&existing, client.Id).Error; err != nil {
		return
	}
	if !up {
		client.UpMbps = existing.UpMbps
	}
	if !down {
		client.DownMbps = existing.DownMbps
	}
}

func (s *ClientService) preserveMissingRatesBulk(tx *gorm.DB, clients []*model.Client, data json.RawMessage) {
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil || len(raws) != len(clients) {
		return
	}
	type presence struct{ up, down bool }
	pres := make(map[uint]presence, len(clients))
	missing := make([]uint, 0, len(clients))
	for i, client := range clients {
		up, down := clientRateFieldsPresent(raws[i])
		if up && down {
			continue
		}
		pres[client.Id] = presence{up: up, down: down}
		missing = append(missing, client.Id)
	}
	if len(missing) == 0 {
		return
	}
	var rows []model.Client
	if err := tx.Select("id", "up_mbps", "down_mbps").Where("id IN ?", missing).Find(&rows).Error; err != nil {
		return
	}
	old := make(map[uint]model.Client, len(rows))
	for _, row := range rows {
		old[row.Id] = row
	}
	for _, client := range clients {
		p, ok := pres[client.Id]
		if !ok {
			continue
		}
		existing, ok := old[client.Id]
		if !ok {
			continue
		}
		if !p.up {
			client.UpMbps = existing.UpMbps
		}
		if !p.down {
			client.DownMbps = existing.DownMbps
		}
	}
}

// validateClientName rejects empty names and globally duplicate names,
// then stores the trimmed name back on the client.
func (s *ClientService) validateClientName(tx *gorm.DB, client *model.Client) error {
	name := strings.TrimSpace(client.Name)
	if name == "" {
		return common.NewError("client name must not be empty")
	}
	query := tx.Model(model.Client{}).Where("name = ?", name)
	if client.Id != 0 {
		query = query.Where("id != ?", client.Id)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return common.NewErrorf("client name %q is already in use", name)
	}
	client.Name = name
	return nil
}

func (s *ClientService) Save(tx *gorm.DB, act string, data json.RawMessage, hostname string) ([]uint, error) {
	var err error
	var inboundIds []uint

	switch act {
	case "new", "edit":
		var client model.Client
		err = json.Unmarshal(data, &client)
		if err != nil {
			return nil, err
		}
		if err = s.validateClientName(tx, &client); err != nil {
			return nil, err
		}
		if act == "edit" {
			s.preserveMissingRates(tx, &client, data)
		}
		if client.MaxConns < 0 {
			client.MaxConns = 0
		}
		if client.IPLimit < 0 {
			client.IPLimit = 0
		}
		if act == "edit" {
			s.preserveMissingIPLimit(tx, &client, data)
		}
		if err = validateClientRates(&client); err != nil {
			return nil, err
		}
		if err = setConfigIdentity(&client); err != nil {
			return nil, err
		}
		err = s.updateLinksWithFixedInbounds(tx, []*model.Client{&client}, hostname)
		if err != nil {
			return nil, err
		}
		if act == "edit" {
			// Find changed inbounds
			inboundIds, err = s.findInboundsChanges(tx, &client, false)
			if err != nil {
				return nil, err
			}
			// Preserve managed timestamps (immutable createdAt, stats-managed onlineAt)
			s.preserveServerOwnedFields(tx, &client)
		} else {
			client.CreatedAt = time.Now().Unix()
			err = json.Unmarshal(client.Inbounds, &inboundIds)
			if err != nil {
				return nil, err
			}
		}
		err = tx.Save(&client).Error
		if err != nil {
			return nil, err
		}
	case "addbulk":
		var clients []*model.Client
		err = json.Unmarshal(data, &clients)
		if err != nil {
			return nil, err
		}
		now := time.Now().Unix()
		// Every client is validated before any of them is written, so the
		// batch has to be checked against itself as well as against the table.
		seen := make(map[string]bool, len(clients))
		for _, client := range clients {
			if err = s.validateClientName(tx, client); err != nil {
				return nil, err
			}
			if client.MaxConns < 0 {
				client.MaxConns = 0
			}
			if client.IPLimit < 0 {
				client.IPLimit = 0
			}
			if err = validateClientRates(client); err != nil {
				return nil, err
			}
			if seen[client.Name] {
				return nil, common.NewErrorf("duplicate client name %q in request", client.Name)
			}
			seen[client.Name] = true
			if err = setConfigIdentity(client); err != nil {
				return nil, err
			}
			client.CreatedAt = now
			var ids []uint
			if err = json.Unmarshal(client.Inbounds, &ids); err != nil {
				return nil, err
			}
			inboundIds = common.UnionUintArray(inboundIds, ids)
		}
		err = s.updateLinksWithFixedInbounds(tx, clients, hostname)
		if err != nil {
			return nil, err
		}
		err = tx.Save(clients).Error
		if err != nil {
			return nil, err
		}
	case "editbulk":
		var clients []*model.Client
		err = json.Unmarshal(data, &clients)
		if err != nil {
			return nil, err
		}
		s.preserveMissingRatesBulk(tx, clients, data)
		seen := make(map[string]bool, len(clients))
		for _, client := range clients {
			if err = s.validateClientName(tx, client); err != nil {
				return nil, err
			}
			if client.MaxConns < 0 {
				client.MaxConns = 0
			}
			if client.IPLimit < 0 {
				client.IPLimit = 0
			}
			if err = validateClientRates(client); err != nil {
				return nil, err
			}
			if seen[client.Name] {
				return nil, common.NewErrorf("duplicate client name %q in request", client.Name)
			}
			seen[client.Name] = true
			changedInboundIds, err := s.findInboundsChanges(tx, client, true)
			if err != nil {
				return nil, err
			}
			if err = setConfigIdentity(client); err != nil {
				return nil, err
			}
			s.preserveServerOwnedFields(tx, client)
			if len(changedInboundIds) > 0 {
				inboundIds = common.UnionUintArray(inboundIds, changedInboundIds)
			}
		}
		if len(inboundIds) > 0 {
			err = s.updateLinksWithFixedInbounds(tx, clients, hostname)
			if err != nil {
				return nil, err
			}
		}
		err = tx.Save(clients).Error
		if err != nil {
			return nil, err
		}
	default:
		return nil, common.NewErrorf("unknown action: %s", act)
	}

	return inboundIds, nil
}

// clientDeleteIDs normalizes the single-id and bulk-id payloads used by the
// clients delete actions.
func clientDeleteIDs(act string, data json.RawMessage) ([]uint, error) {
	if act == "del" {
		var id uint
		if err := json.Unmarshal(data, &id); err != nil {
			return nil, err
		}
		return []uint{id}, nil
	}
	var ids []uint
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// DeleteClients removes clients together with the outbounds and top-level
// auth_user routes that were generated for them. A route is only removed when
// every auth_user it names is being deleted and its outbound is the generated
// out-<client> tag; shared or conditional rules are left untouched. The second
// return value reports whether the running core must be restarted because its
// route/outbound set changed.
func (s *ClientService) DeleteClients(tx *gorm.DB, ids []uint) ([]uint, bool, error) {
	var inboundIds []uint
	if len(ids) == 0 {
		return inboundIds, false, nil
	}
	// Keep the IN list below SQLite's bound-variable limit; a select-all across
	// many pages can legitimately exceed it.
	const deleteChunk = 400
	var clients []model.Client
	for start := 0; start < len(ids); start += deleteChunk {
		end := start + deleteChunk
		if end > len(ids) {
			end = len(ids)
		}
		var batch []model.Client
		if err := tx.Where("id in ?", ids[start:end]).Find(&batch).Error; err != nil {
			return nil, false, err
		}
		clients = append(clients, batch...)
	}
	if len(clients) == 0 {
		return inboundIds, false, nil
	}
	names := make(map[string]bool, len(clients))
	for _, client := range clients {
		names[client.Name] = true
		var clientInbounds []uint
		if err := json.Unmarshal(client.Inbounds, &clientInbounds); err != nil {
			return nil, false, err
		}
		inboundIds = common.UnionUintArray(inboundIds, clientInbounds)
	}

	changed, err := s.removeClientRoutes(tx, names)
	if err != nil {
		return nil, false, err
	}
	for start := 0; start < len(ids); start += deleteChunk {
		end := start + deleteChunk
		if end > len(ids) {
			end = len(ids)
		}
		if err = tx.Where("id in ?", ids[start:end]).Delete(model.Client{}).Error; err != nil {
			return nil, false, err
		}
	}
	if err = DeleteClientSubscriptions(tx, ids); err != nil {
		return nil, false, err
	}
	return inboundIds, changed, nil
}

func authUserNames(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var names []string
	if err := json.Unmarshal(raw, &names); err == nil {
		return names
	}
	var name string
	if err := json.Unmarshal(raw, &name); err == nil && name != "" {
		return []string{name}
	}
	return nil
}

// removeClientRoutes deletes the generated route/outbound pair for each name.
// The stored config is updated in the same transaction as the outbound rows.
func (s *ClientService) removeClientRoutes(tx *gorm.DB, names map[string]bool) (bool, error) {
	var setting model.Setting
	if err := tx.Where("key = ?", "config").Limit(1).Find(&setting).Error; err != nil {
		return false, err
	}
	if strings.TrimSpace(setting.Value) == "" {
		return false, nil
	}
	var cfg SingBoxConfig
	if err := json.Unmarshal([]byte(setting.Value), &cfg); err != nil {
		return false, common.NewError("cannot read the stored route config: ", err)
	}
	route := map[string]json.RawMessage{}
	if len(cfg.Route) > 0 {
		if err := json.Unmarshal(cfg.Route, &route); err != nil {
			return false, common.NewError("cannot read route: ", err)
		}
	}
	var rules []json.RawMessage
	if raw, ok := route["rules"]; ok && len(raw) > 0 {
		if err := json.Unmarshal(raw, &rules); err != nil {
			return false, common.NewError("cannot read route rules: ", err)
		}
	}

	removedTags := make(map[string]bool)
	removedRule := false
	kept := make([]json.RawMessage, 0, len(rules))
	for _, rule := range rules {
		var probe struct {
			AuthUser json.RawMessage `json:"auth_user"`
			Action   string          `json:"action"`
			Outbound string          `json:"outbound"`
			Type     string          `json:"type"`
			Invert   bool            `json:"invert"`
		}
		if err := json.Unmarshal(rule, &probe); err != nil {
			kept = append(kept, rule)
			continue
		}
		users := authUserNames(probe.AuthUser)
		// Only simple top-level auth_user rules are safe to delete. Logical or
		// inverted rules can affect other users, so they stay untouched.
		if len(users) == 0 || probe.Type == "logical" || probe.Invert {
			kept = append(kept, rule)
			continue
		}
		allDeleted := true
		for _, user := range users {
			if !names[user] {
				allDeleted = false
				break
			}
		}
		if !allDeleted {
			kept = append(kept, rule)
			continue
		}
		if probe.Outbound != "" {
			removedTags[probe.Outbound] = true
		}
		removedRule = true
	}
	if !removedRule {
		return false, nil
	}

	// Keep an outbound that is still the route.final or referenced by a
	// surviving rule; deleting it would make the stored config invalid.
	referenced := make(map[string]bool)
	for _, rule := range kept {
		var probe struct {
			Outbound string `json:"outbound"`
		}
		if json.Unmarshal(rule, &probe) == nil && probe.Outbound != "" {
			referenced[probe.Outbound] = true
		}
	}
	referenced[routeFinal(route)] = true
	for tag := range removedTags {
		if referenced[tag] {
			delete(removedTags, tag)
		}
	}

	encodedRules, err := json.Marshal(kept)
	if err != nil {
		return false, err
	}
	route["rules"] = encodedRules
	encodedRoute, err := json.Marshal(route)
	if err != nil {
		return false, err
	}
	cfg.Route = encodedRoute
	encodedConfig, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, err
	}
	if err = tx.Model(model.Setting{}).Where("key = ?", "config").Update("value", string(encodedConfig)).Error; err != nil {
		return false, err
	}
	for tag := range removedTags {
		var outbound model.Outbound
		err := tx.Where("tag = ?", tag).First(&outbound).Error
		if err == gorm.ErrRecordNotFound {
			continue
		}
		if err != nil {
			return false, err
		}
		// Built-in outbounds are shared defaults, not per-client resources.
		if outbound.Type == "direct" || outbound.Type == "block" {
			continue
		}
		if err = tx.Where("tag = ?", tag).Delete(model.Outbound{}).Error; err != nil {
			return false, err
		}
	}
	return true, nil
}

func routeFinal(route map[string]json.RawMessage) string {
	raw, ok := route["final"]
	if !ok {
		return ""
	}
	var final string
	if json.Unmarshal(raw, &final) != nil {
		return ""
	}
	return final
}

// preserveServerOwnedFields restores the columns the panel maintains itself.
// The traffic counters matter as much as the timestamps: the stats job writes
// up/down every ten seconds, so a stale form would roll them back.
func (s *ClientService) preserveServerOwnedFields(tx *gorm.DB, client *model.Client) {
	var existing model.Client
	if err := tx.Model(model.Client{}).
		Select("created_at", "online_at", "up", "down", "total_up", "total_down").
		Where("id = ?", client.Id).First(&existing).Error; err != nil {
		return
	}
	client.CreatedAt = existing.CreatedAt
	client.OnlineAt = existing.OnlineAt

	if client.Up == 0 && client.Down == 0 {
		client.TotalUp = existing.TotalUp + existing.Up
		client.TotalDown = existing.TotalDown + existing.Down
		return
	}

	client.Up = existing.Up
	client.Down = existing.Down
	client.TotalUp = existing.TotalUp
	client.TotalDown = existing.TotalDown
}

func linkRemark(prefix bool, remark string) string {
	if !prefix {
		return ""
	}
	return remark
}

func clientNameJSON(name string) json.RawMessage {
	encoded, err := json.Marshal(name)
	if err != nil {
		// json.Marshal of a string cannot fail, but never emit broken JSON.
		return json.RawMessage(`""`)
	}
	return json.RawMessage(encoded)
}

func (s *ClientService) updateLinksWithFixedInbounds(tx *gorm.DB, clients []*model.Client, hostname string) error {
	remarkPrefix := (&SettingService{}).SubRemarkPrefix(tx)
	clientInboundIds := make([][]uint, len(clients))
	var allIds []uint
	for i, client := range clients {
		var ids []uint
		if err := json.Unmarshal(client.Inbounds, &ids); err != nil {
			return err
		}
		clientInboundIds[i] = ids
		allIds = common.UnionUintArray(allIds, ids)
	}

	// Zero inbounds means removing local links only
	var inbounds []model.Inbound
	if len(allIds) > 0 {
		err := tx.Model(model.Inbound{}).Preload("Tls").Where("id in ? and type in ?", allIds, util.InboundTypeWithLink).Find(&inbounds).Error
		if err != nil {
			return err
		}
	}
	inboundById := make(map[uint]*model.Inbound, len(inbounds))
	for i := range inbounds {
		inboundById[inbounds[i].Id] = &inbounds[i]
	}

	for index, client := range clients {
		var clientLinks []map[string]string
		if err := json.Unmarshal(client.Links, &clientLinks); err != nil {
			return err
		}

		newClientLinks := []map[string]string{}
		for _, id := range clientInboundIds[index] {
			inbound, ok := inboundById[id]
			if !ok {
				continue
			}
			newLinks := util.LinkGenerator(client.Config, inbound, hostname, linkRemark(remarkPrefix, client.Remark))
			for _, newLink := range newLinks {
				newClientLinks = append(newClientLinks, map[string]string{
					"remark": inbound.Tag,
					"type":   "local",
					"uri":    newLink,
				})
			}
		}

		// Add non local links
		for _, clientLink := range clientLinks {
			if clientLink["type"] != "local" {
				newClientLinks = append(newClientLinks, clientLink)
			}
		}

		links, err := json.MarshalIndent(newClientLinks, "", "  ")
		if err != nil {
			return err
		}
		clients[index].Links = links
	}
	return nil
}

func (s *ClientService) UpdateClientsOnInboundAdd(tx *gorm.DB, initIds string, inboundId uint, hostname string) error {
	remarkPrefix := (&SettingService{}).SubRemarkPrefix(tx)
	clientIds := strings.Split(initIds, ",")
	var clients []model.Client
	err := tx.Model(model.Client{}).Where("id in ?", clientIds).Find(&clients).Error
	if err != nil {
		return err
	}
	var inbound model.Inbound
	err = tx.Model(model.Inbound{}).Preload("Tls").Where("id = ?", inboundId).Find(&inbound).Error
	if err != nil {
		return err
	}
	for _, client := range clients {
		// Add inbounds
		var clientInbounds []uint
		json.Unmarshal(client.Inbounds, &clientInbounds)
		clientInbounds = append(clientInbounds, inboundId)
		client.Inbounds, err = json.MarshalIndent(clientInbounds, "", "  ")
		if err != nil {
			return err
		}
		// Add links
		var clientLinks, newClientLinks []map[string]string
		json.Unmarshal(client.Links, &clientLinks)
		newLinks := util.LinkGenerator(client.Config, &inbound, hostname, linkRemark(remarkPrefix, client.Remark))
		for _, newLink := range newLinks {
			newClientLinks = append(newClientLinks, map[string]string{
				"remark": inbound.Tag,
				"type":   "local",
				"uri":    newLink,
			})
		}
		for _, clientLink := range clientLinks {
			if clientLink["remark"] != inbound.Tag {
				newClientLinks = append(newClientLinks, clientLink)
			}
		}

		client.Links, err = json.MarshalIndent(newClientLinks, "", "  ")
		if err != nil {
			return err
		}
		err = tx.Save(&client).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ClientService) UpdateClientsOnInboundDelete(tx *gorm.DB, id uint, tag string) error {
	var clientIds []uint
	err := tx.Raw("SELECT clients.id FROM clients, json_each(CAST(clients.inbounds AS TEXT)) AS je WHERE je.value = ?", id).Scan(&clientIds).Error
	if err != nil {
		return err
	}
	if len(clientIds) == 0 {
		return nil
	}
	var clients []model.Client
	err = tx.Model(model.Client{}).Where("id IN ?", clientIds).Find(&clients).Error
	if err != nil {
		return err
	}
	for _, client := range clients {
		// Delete inbounds
		var clientInbounds, newClientInbounds []uint
		json.Unmarshal(client.Inbounds, &clientInbounds)
		for _, clientInbound := range clientInbounds {
			if clientInbound != id {
				newClientInbounds = append(newClientInbounds, clientInbound)
			}
		}
		client.Inbounds, err = json.MarshalIndent(newClientInbounds, "", "  ")
		if err != nil {
			return err
		}
		// Delete links
		var clientLinks, newClientLinks []map[string]string
		json.Unmarshal(client.Links, &clientLinks)
		for _, clientLink := range clientLinks {
			if clientLink["remark"] != tag {
				newClientLinks = append(newClientLinks, clientLink)
			}
		}
		client.Links, err = json.MarshalIndent(newClientLinks, "", "  ")
		if err != nil {
			return err
		}
		err = tx.Save(&client).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ClientService) UpdateLinksByInboundChange(tx *gorm.DB, inbounds *[]model.Inbound, hostname string, oldTag string) error {
	remarkPrefix := (&SettingService{}).SubRemarkPrefix(tx)
	var err error
	for _, inbound := range *inbounds {
		var clientIds []uint
		err = tx.Raw("SELECT clients.id FROM clients, json_each(CAST(clients.inbounds AS TEXT)) AS je WHERE je.value = ?", inbound.Id).Scan(&clientIds).Error
		if err != nil {
			return err
		}
		if len(clientIds) == 0 {
			continue
		}
		var clients []model.Client
		err = tx.Model(model.Client{}).Where("id IN ?", clientIds).Find(&clients).Error
		if err != nil {
			return err
		}
		for _, client := range clients {
			var clientLinks, newClientLinks []map[string]string
			json.Unmarshal(client.Links, &clientLinks)
			newLinks := util.LinkGenerator(client.Config, &inbound, hostname, linkRemark(remarkPrefix, client.Remark))
			for _, newLink := range newLinks {
				newClientLinks = append(newClientLinks, map[string]string{
					"remark": inbound.Tag,
					"type":   "local",
					"uri":    newLink,
				})
			}
			for _, clientLink := range clientLinks {
				if clientLink["type"] != "local" || (clientLink["remark"] != inbound.Tag && clientLink["remark"] != oldTag) {
					newClientLinks = append(newClientLinks, clientLink)
				}
			}

			client.Links, err = json.MarshalIndent(newClientLinks, "", "  ")
			if err != nil {
				return err
			}
			err = tx.Save(&client).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ClientService) DepleteClients() ([]uint, error) {
	var err error
	var clients []model.Client
	var changes []model.Changes
	var users []string
	var inboundIds []uint

	dt := time.Now().Unix()
	db := database.GetDB()

	tx := db.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
			if err1 := db.Exec("PRAGMA wal_checkpoint(FULL)").Error; err1 != nil {
				logger.Error("Error checkpointing WAL: ", err1.Error())
			}
		} else {
			tx.Rollback()
		}
	}()

	// Reset clients
	inboundIds, err = s.ResetClients(tx, dt)
	if err != nil {
		return nil, err
	}

	// Deplete clients
	err = tx.Model(model.Client{}).Where("enable = true AND ((volume >0 AND up+down > volume) OR (expiry > 0 AND expiry < ?))", dt).Scan(&clients).Error
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		logger.Debug("Client ", client.Name, " is going to be disabled")
		users = append(users, client.Name)
		var userInbounds []uint
		json.Unmarshal(client.Inbounds, &userInbounds)
		// Find changed inbounds
		inboundIds = common.UnionUintArray(inboundIds, userInbounds)
		changes = append(changes, model.Changes{
			DateTime: dt,
			Actor:    "DepleteJob",
			Key:      "clients",
			Action:   "disable",
			Obj:      clientNameJSON(client.Name),
		})
	}

	// Save changes
	if len(changes) > 0 {
		err = tx.Model(model.Client{}).Where("enable = true AND ((volume >0 AND up+down > volume) OR (expiry > 0 AND expiry < ?))", dt).Update("enable", false).Error
		if err != nil {
			return nil, err
		}
		err = tx.Model(model.Changes{}).Create(&changes).Error
		if err != nil {
			return nil, err
		}
		LastUpdate = dt
	}

	return inboundIds, nil
}

func (s *ClientService) ResetClients(tx *gorm.DB, dt int64) ([]uint, error) {
	var err error
	var resetClients, allClients []*model.Client
	var changes []model.Changes
	var inboundIds []uint
	// Set delay start without periodic reset
	err = tx.Model(model.Client{}).
		Where("enable = true AND delay_start = true AND auto_reset = false AND (Up + Down) > 0").Find(&resetClients).Error
	if err != nil {
		return nil, err
	}
	for _, client := range resetClients {
		client.Expiry = dt + (int64(client.ResetDays) * 86400)
		client.DelayStart = false
		changes = append(changes, model.Changes{
			DateTime: dt,
			Actor:    "ResetJob",
			Key:      "clients",
			Action:   "reset",
			Obj:      clientNameJSON(client.Name),
		})
	}
	allClients = append(allClients, resetClients...)

	// Set delay start with periodic reset
	err = tx.Model(model.Client{}).
		Where("enable = true AND delay_start = true AND auto_reset = true AND (Up + Down) > 0").Find(&resetClients).Error
	if err != nil {
		return nil, err
	}
	for _, client := range resetClients {
		client.NextReset = dt + (int64(client.ResetDays) * 86400)
		client.DelayStart = false
		changes = append(changes, model.Changes{
			DateTime: dt,
			Actor:    "ResetJob",
			Key:      "clients",
			Action:   "reset",
			Obj:      clientNameJSON(client.Name),
		})
	}
	allClients = append(allClients, resetClients...)

	// reset_days > 0 is a backstop: at zero, NextReset becomes dt + 0 == dt, so
	// the row matches every minute and the volume quota is never reached.
	err = tx.Model(model.Client{}).
		Where("delay_start = false AND auto_reset = true AND reset_days > 0 AND next_reset < ?", dt).Find(&resetClients).Error
	if err != nil {
		return nil, err
	}
	for _, client := range resetClients {
		client.NextReset = dt + (int64(client.ResetDays) * 86400)
		client.TotalUp += client.Up
		client.TotalDown += client.Down
		client.Up = 0
		client.Down = 0
		changes = append(changes, model.Changes{
			DateTime: dt,
			Actor:    "ResetJob",
			Key:      "clients",
			Action:   "reset",
			Obj:      clientNameJSON(client.Name),
		})
		if !client.Enable {
			client.Enable = true
			var clientInboundIds []uint
			json.Unmarshal(client.Inbounds, &clientInboundIds)
			inboundIds = common.UnionUintArray(inboundIds, clientInboundIds)
		}
	}
	allClients = append(allClients, resetClients...)

	// Save clients
	if len(allClients) > 0 {
		err = tx.Save(allClients).Error
		if err != nil {
			return nil, err
		}
	}

	// Save changes
	if len(changes) > 0 {
		err = tx.Model(model.Changes{}).Create(&changes).Error
		if err != nil {
			return nil, err
		}
		LastUpdate = dt
	}
	return inboundIds, nil
}

// ResetAllClientsTraffic zeroes up/down for every client (accumulating into the
// total counters) and re-enables all of them, in a single bulk update. Used by
// the global periodic traffic reset. Returns the inbounds of the clients it
// re-enabled, whose users the caller updates in the running core.
func (s *ClientService) ResetAllClientsTraffic() ([]uint, error) {
	db := database.GetDB()
	dt := time.Now().Unix()
	var inboundIds []uint

	err := db.Transaction(func(tx *gorm.DB) error {
		var disabled []model.Client
		if err := tx.Model(model.Client{}).Select("inbounds").Where("enable = false").Find(&disabled).Error; err != nil {
			return err
		}
		for _, client := range disabled {
			var userInbounds []uint
			json.Unmarshal(client.Inbounds, &userInbounds)
			inboundIds = common.UnionUintArray(inboundIds, userInbounds)
		}

		result := tx.Model(model.Client{}).
			Where("(up + down) > 0 OR enable = false").
			UpdateColumns(map[string]interface{}{
				"total_up":   gorm.Expr("total_up + up"),
				"total_down": gorm.Expr("total_down + down"),
				"up":         0,
				"down":       0,
				"enable":     true,
			})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected > 0 {
			if err := tx.Create(&model.Changes{
				DateTime: dt,
				Actor:    "ResetTrafficJob",
				Key:      "clients",
				Action:   "reset",
				Obj:      json.RawMessage("\"all\""),
			}).Error; err != nil {
				return err
			}
			LastUpdate = dt
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inboundIds, nil
}

func setConfigIdentity(client *model.Client) error {
	if client.Name == "" || len(client.Config) < 2 {
		return nil
	}
	var configs map[string]map[string]interface{}
	if err := json.Unmarshal(client.Config, &configs); err != nil {
		return err
	}
	for _, cfg := range configs {
		if _, ok := cfg["name"]; ok {
			cfg["name"] = client.Name
		} else if _, ok := cfg["username"]; ok {
			cfg["username"] = client.Name
		}
	}
	newConfig, err := json.Marshal(configs)
	if err != nil {
		return err
	}
	client.Config = newConfig
	return nil
}

func (s *ClientService) findInboundsChanges(tx *gorm.DB, client *model.Client, fillOmitted bool) ([]uint, error) {
	var err error
	var oldClient model.Client
	var oldInboundIds, newInboundIds []uint
	err = tx.Model(model.Client{}).Where("id = ?", client.Id).First(&oldClient).Error
	if err != nil {
		return nil, err
	}
	if fillOmitted {
		client.Links = oldClient.Links
		client.Config = oldClient.Config
	}
	err = json.Unmarshal(oldClient.Inbounds, &oldInboundIds)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(client.Inbounds, &newInboundIds)
	if err != nil {
		return nil, err
	}

	// Check client.Config changes
	if !bytes.Equal(oldClient.Config, client.Config) ||
		oldClient.Name != client.Name ||
		oldClient.Enable != client.Enable {
		return common.UnionUintArray(oldInboundIds, newInboundIds), nil
	}

	// Check client.Inbounds changes
	diffInbounds := common.DiffUintArray(oldInboundIds, newInboundIds)

	return diffInbounds, nil
}
