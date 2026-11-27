package service

import (
	"encoding/json"
	"testing"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"

	"gorm.io/gorm"
)

func seedDeleteTest(t *testing.T, db *gorm.DB, config string, tags ...string) (uint, uint) {
	t.Helper()
	if err := db.Create(&model.Setting{Key: "config", Value: config}).Error; err != nil {
		t.Fatal(err)
	}
	for _, tag := range tags {
		var existing model.Outbound
		err := db.Where("tag = ?", tag).First(&existing).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			t.Fatal(err)
		}
		if err := db.Create(&model.Outbound{Type: "socks", Tag: tag, Options: json.RawMessage(`{"server":"1.1.1.1","server_port":80}`)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	alice := createClient(t, db, &model.Client{Name: "alice", Enable: true, Volume: 100})
	bob := createClient(t, db, &model.Client{Name: "bob", Enable: true, Volume: 100})
	return alice.Id, bob.Id
}

func deleteTestClients(t *testing.T, ids ...uint) bool {
	t.Helper()
	db := database.GetDB()
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	_, changed, err := (&ClientService{}).DeleteClients(tx, ids)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	return changed
}

func outboundExists(t *testing.T, db *gorm.DB, tag string) bool {
	t.Helper()
	var count int64
	if err := db.Model(model.Outbound{}).Where("tag = ?", tag).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	return count > 0
}

func routeUsers(t *testing.T, db *gorm.DB) [][]string {
	t.Helper()
	var setting model.Setting
	if err := db.Where("key = ?", "config").First(&setting).Error; err != nil {
		t.Fatal(err)
	}
	var cfg SingBoxConfig
	if err := json.Unmarshal([]byte(setting.Value), &cfg); err != nil {
		t.Fatal(err)
	}
	var route struct {
		Rules []json.RawMessage `json:"rules"`
	}
	if err := json.Unmarshal(cfg.Route, &route); err != nil {
		t.Fatal(err)
	}
	result := make([][]string, 0, len(route.Rules))
	for _, raw := range route.Rules {
		var probe struct {
			AuthUser json.RawMessage `json:"auth_user"`
		}
		if err := json.Unmarshal(raw, &probe); err != nil {
			continue
		}
		result = append(result, authUserNames(probe.AuthUser))
	}
	return result
}

func TestDeleteClientsRemovesGeneratedOutboundAndRoute(t *testing.T) {
	db := clientTestDB(t)
	alice, _ := seedDeleteTest(t, db, `{
		"outbounds": [],
		"route": {"final": "direct", "rules": [
			{"action": "sniff"},
			{"auth_user": ["alice"], "action": "route", "outbound": "out-alice"},
			{"auth_user": ["bob"], "action": "route", "outbound": "out-bob"}
		]}
	}`, "out-alice", "out-bob", "direct")

	if !deleteTestClients(t, alice) {
		t.Fatal("changed = false, want true for a removed route")
	}
	var count int64
	db.Model(model.Client{}).Where("id = ?", alice).Count(&count)
	if count != 0 {
		t.Fatal("alice was not deleted")
	}
	if outboundExists(t, db, "out-alice") {
		t.Fatal("out-alice was not deleted")
	}
	if !outboundExists(t, db, "out-bob") {
		t.Fatal("out-bob must remain while bob exists")
	}
	for _, users := range routeUsers(t, db) {
		for _, user := range users {
			if user == "alice" {
				t.Fatal("alice route was not removed")
			}
		}
	}
}

func TestDeleteClientsKeepsSharedOutboundAndConditionalRoute(t *testing.T) {
	db := clientTestDB(t)
	alice, _ := seedDeleteTest(t, db, `{
		"outbounds": [],
		"route": {"final": "direct", "rules": [
			{"auth_user": ["alice", "bob"], "action": "route", "outbound": "out-alice"},
			{"domain_suffix": ["example.com"], "action": "route", "outbound": "out-alice"}
		]}
	}`, "out-alice", "direct")

	if deleteTestClients(t, alice) {
		t.Fatal("changed = true, want false when the route/outbound remains in use")
	}
	if !outboundExists(t, db, "out-alice") {
		t.Fatal("shared out-alice was deleted")
	}
	found := false
	for _, users := range routeUsers(t, db) {
		if len(users) == 2 {
			found = true
		}
	}
	if !found {
		t.Fatal("shared auth_user route was removed")
	}
}

func TestDeleteClientsKeepsFinalOutbound(t *testing.T) {
	db := clientTestDB(t)
	alice, _ := seedDeleteTest(t, db, `{
		"outbounds": [],
		"route": {"final": "out-alice", "rules": [
			{"auth_user": ["alice"], "action": "route", "outbound": "out-alice"}
		]}
	}`, "out-alice")

	if !deleteTestClients(t, alice) {
		t.Fatal("changed = false, want true for the removed route")
	}
	if !outboundExists(t, db, "out-alice") {
		t.Fatal("final out-alice was deleted, leaving an invalid route.final")
	}
}

func TestDeleteClientsBatchRemovesBothGeneratedPairs(t *testing.T) {
	db := clientTestDB(t)
	alice, bob := seedDeleteTest(t, db, `{
		"outbounds": [],
		"route": {"final": "direct", "rules": [
			{"auth_user": ["alice"], "action": "route", "outbound": "out-alice"},
			{"auth_user": ["bob"], "action": "route", "outbound": "out-bob"}
		]}
	}`, "out-alice", "out-bob", "direct")

	if !deleteTestClients(t, alice, bob) {
		t.Fatal("changed = false for a batch delete")
	}
	var count int64
	db.Model(model.Client{}).Where("id in ?", []uint{alice, bob}).Count(&count)
	if count != 0 {
		t.Fatalf("clients left = %d", count)
	}
	for _, tag := range []string{"out-alice", "out-bob"} {
		if outboundExists(t, db, tag) {
			t.Fatalf("%s was not deleted", tag)
		}
	}
	if users := routeUsers(t, db); len(users) != 0 {
		t.Fatalf("routes left = %v", users)
	}
}

func TestDeleteClientsChunksLargeIDLists(t *testing.T) {
	db := clientTestDB(t)
	clients := make([]model.Client, 0, 401)
	for i := 0; i < 401; i++ {
		clients = append(clients, model.Client{
			Name:     "many-" + string(rune('a'+i%26)) + "-" + string(rune('a'+i/26)),
			Enable:   true,
			Config:   json.RawMessage(`{}`),
			Inbounds: json.RawMessage(`[]`),
			Links:    json.RawMessage(`[]`),
		})
	}
	if err := db.Create(&clients).Error; err != nil {
		t.Fatal(err)
	}
	ids := make([]uint, 0, len(clients))
	for i := range clients {
		ids = append(ids, clients[i].Id)
	}
	if deleteTestClients(t, ids...) {
		t.Fatal("changed = true without routes or outbounds")
	}
	var count int64
	db.Model(model.Client{}).Where("id in ?", ids[:400]).Count(&count)
	if count != 0 {
		t.Fatalf("clients left = %d", count)
	}
}

func TestDeleteClientsRemovesRouteWithCustomOutboundTag(t *testing.T) {
	db := clientTestDB(t)
	alice, _ := seedDeleteTest(t, db, `{
		"outbounds": [],
		"route": {"final": "direct", "rules": [
			{"auth_user": ["alice"], "action": "route", "outbound": "socks-s28"}
		]}
	}`, "socks-s28", "direct")

	if !deleteTestClients(t, alice) {
		t.Fatal("changed = false for a custom-tag route")
	}
	if outboundExists(t, db, "socks-s28") {
		t.Fatal("socks-s28 was not deleted")
	}
	if !outboundExists(t, db, "direct") {
		t.Fatal("built-in direct was deleted")
	}
	for _, users := range routeUsers(t, db) {
		for _, user := range users {
			if user == "alice" {
				t.Fatal("alice custom-tag route was not removed")
			}
		}
	}
}

func TestDeleteClientsKeepsDirectWhenItsRuleIsRemoved(t *testing.T) {
	db := clientTestDB(t)
	alice, _ := seedDeleteTest(t, db, `{
		"outbounds": [],
		"route": {"final": "direct", "rules": [
			{"auth_user": ["alice"], "action": "route", "outbound": "direct"}
		]}
	}`)

	if !deleteTestClients(t, alice) {
		t.Fatal("changed = false for a direct route")
	}
	if !outboundExists(t, db, "direct") {
		t.Fatal("direct was deleted")
	}
}

func TestDeleteClientsRemovesPersonalSubscriptionAndEmptyGroup(t *testing.T) {
	db := clientTestDB(t)
	alice, _ := seedDeleteTest(t, db, `{
		"outbounds": [],
		"route": {"final": "direct", "rules": []}
	}`)
	group := model.SubscriptionGroup{SubID: "group-alone", Name: "alone"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	member := model.SubscriptionMember{GroupId: group.Id, ClientId: alice, SubID: "member-alone"}
	if err := db.Create(&member).Error; err != nil {
		t.Fatal(err)
	}
	if deleteTestClients(t, alice) {
		t.Fatal("changed = true without a route")
	}
	var members, groups int64
	db.Model(model.SubscriptionMember{}).Where("client_id = ?", alice).Count(&members)
	db.Model(model.SubscriptionGroup{}).Where("id = ?", group.Id).Count(&groups)
	if members != 0 || groups != 0 {
		t.Fatalf("subscription rows left: members=%d groups=%d", members, groups)
	}
}

func TestDeleteClientsKeepsSharedSubscriptionGroup(t *testing.T) {
	db := clientTestDB(t)
	alice, bob := seedDeleteTest(t, db, `{
		"outbounds": [],
		"route": {"final": "direct", "rules": []}
	}`)
	group := model.SubscriptionGroup{SubID: "group-shared", Name: "shared"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range []model.SubscriptionMember{
		{GroupId: group.Id, ClientId: alice, SubID: "member-a"},
		{GroupId: group.Id, ClientId: bob, SubID: "member-b"},
	} {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	if deleteTestClients(t, alice) {
		t.Fatal("changed = true without a route")
	}
	var members, groups int64
	db.Model(model.SubscriptionMember{}).Where("client_id = ?", alice).Count(&members)
	db.Model(model.SubscriptionGroup{}).Where("id = ?", group.Id).Count(&groups)
	if members != 0 || groups != 1 {
		t.Fatalf("members=%d groups=%d, want 0/1", members, groups)
	}
}
