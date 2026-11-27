package service

import (
	"encoding/json"
	"testing"

	"github.com/alireza0/s-ui/database/model"
)

func TestReplaceServersPreservesOtherOptions(t *testing.T) {
	db := clientTestDB(t)
	previous := corePtr
	corePtr = nil
	t.Cleanup(func() { corePtr = previous })

	socks := model.Outbound{
		Type: "socks",
		Tag:  "socks-a",
		Options: json.RawMessage(`{
			"server": "192.0.2.1",
			"server_port": 1111,
			"username": "old-a",
			"password": "old-a-pass",
			"version": "5",
			"detour": "direct",
			"udp_over_tcp": true
		}`),
	}
	httpOut := model.Outbound{
		Type: "http",
		Tag:  "http-b",
		Options: json.RawMessage(`{
			"server": "192.0.2.2",
			"server_port": 2222,
			"username": "old-b",
			"password": "old-b-pass",
			"tls": {"enabled": true, "server_name": "proxy.example"}
		}`),
	}
	if err := db.Create(&socks).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&httpOut).Error; err != nil {
		t.Fatal(err)
	}

	result, err := (&OutboundService{}).ReplaceServers([]ReplaceOutboundItem{
		{Tag: "socks-a", Server: "192.0.2.66", ServerPort: 9873, Username: "user-a", Password: "pass:a"},
		{Tag: "http-b", Server: "192.0.2.61", ServerPort: 8080, Username: "user-b", Password: "pass:b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Saved != 2 || result.Applied || result.Warning == "" {
		t.Fatalf("unexpected result: %+v", result)
	}

	var storedSocks model.Outbound
	if err := db.Where("tag = ?", "socks-a").First(&storedSocks).Error; err != nil {
		t.Fatal(err)
	}
	var socksOptions map[string]json.RawMessage
	if err := json.Unmarshal(storedSocks.Options, &socksOptions); err != nil {
		t.Fatal(err)
	}
	var server string
	_ = json.Unmarshal(socksOptions["server"], &server)
	if server != "192.0.2.66" {
		t.Fatalf("server = %q", server)
	}
	var port uint16
	_ = json.Unmarshal(socksOptions["server_port"], &port)
	if port != 9873 {
		t.Fatalf("port = %d", port)
	}
	var version string
	_ = json.Unmarshal(socksOptions["version"], &version)
	if version != "5" {
		t.Fatalf("socks version lost: %q", version)
	}
	if _, ok := socksOptions["udp_over_tcp"]; !ok {
		t.Fatal("socks udp_over_tcp lost")
	}
	if storedSocks.Type != "socks" || storedSocks.Tag != "socks-a" || storedSocks.Id != socks.Id {
		t.Fatalf("identity changed: %+v", storedSocks)
	}

	var storedHTTP model.Outbound
	if err := db.Where("tag = ?", "http-b").First(&storedHTTP).Error; err != nil {
		t.Fatal(err)
	}
	var httpOptions map[string]json.RawMessage
	if err := json.Unmarshal(storedHTTP.Options, &httpOptions); err != nil {
		t.Fatal(err)
	}
	if _, ok := httpOptions["tls"]; !ok {
		t.Fatal("http tls lost")
	}
	var password string
	_ = json.Unmarshal(httpOptions["password"], &password)
	if password != "pass:b" {
		t.Fatalf("password = %q, want pass:b", password)
	}
}

func TestReplaceServersRollsBackOnMissingTarget(t *testing.T) {
	db := clientTestDB(t)
	previous := corePtr
	corePtr = nil
	t.Cleanup(func() { corePtr = previous })

	original := model.Outbound{
		Type:    "socks",
		Tag:     "socks-a",
		Options: json.RawMessage(`{"server":"192.0.2.1","server_port":1080}`),
	}
	if err := db.Create(&original).Error; err != nil {
		t.Fatal(err)
	}
	_, err := (&OutboundService{}).ReplaceServers([]ReplaceOutboundItem{
		{Tag: "socks-a", Server: "192.0.2.66", ServerPort: 9873},
		{Tag: "missing", Server: "192.0.2.67", ServerPort: 1081},
	})
	if err == nil {
		t.Fatal("expected missing target error")
	}
	var stored model.Outbound
	if err := db.Where("tag = ?", "socks-a").First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if string(stored.Options) == "" || !json.Valid(stored.Options) {
		t.Fatalf("invalid options after rollback: %s", stored.Options)
	}
	var options map[string]json.RawMessage
	_ = json.Unmarshal(stored.Options, &options)
	var server string
	_ = json.Unmarshal(options["server"], &server)
	if server != "192.0.2.1" {
		t.Fatalf("partial update survived rollback: server=%q", server)
	}
}

func TestReplaceServersRejectsDuplicateUnsupportedAndEmptyAuth(t *testing.T) {
	db := clientTestDB(t)
	previous := corePtr
	corePtr = nil
	t.Cleanup(func() { corePtr = previous })

	for _, row := range []model.Outbound{
		{Type: "socks", Tag: "socks-a", Options: json.RawMessage(`{"server":"192.0.2.1","server_port":1080}`)},
		{Type: "vless", Tag: "vless-a", Options: json.RawMessage(`{"server":"192.0.2.2","server_port":443}`)},
	} {
		row := row
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}

	if _, err := (&OutboundService{}).ReplaceServers([]ReplaceOutboundItem{
		{Tag: "socks-a", Server: "192.0.2.66", ServerPort: 9873},
		{Tag: "socks-a", Server: "192.0.2.67", ServerPort: 9874},
	}); err == nil {
		t.Fatal("duplicate target accepted")
	}
	if _, err := (&OutboundService{}).ReplaceServers([]ReplaceOutboundItem{
		{Tag: "vless-a", Server: "192.0.2.66", ServerPort: 9873},
	}); err == nil {
		t.Fatal("unsupported type accepted")
	}
	if _, err := (&OutboundService{}).ReplaceServers([]ReplaceOutboundItem{
		{Tag: "socks-a", Server: "192.0.2.66", ServerPort: 0},
	}); err == nil {
		t.Fatal("port 0 accepted")
	}
	if _, err := (&OutboundService{}).ReplaceServers([]ReplaceOutboundItem{
		{Tag: "socks-a", Server: "192.0.2.66", ServerPort: 9873, Username: "", Password: ""},
	}); err != nil {
		t.Fatal(err)
	}
	var stored model.Outbound
	_ = db.Where("tag = ?", "socks-a").First(&stored).Error
	var options map[string]json.RawMessage
	_ = json.Unmarshal(stored.Options, &options)
	if _, ok := options["username"]; ok {
		t.Fatal("empty username was not removed")
	}
	if _, ok := options["password"]; ok {
		t.Fatal("empty password was not removed")
	}
}
