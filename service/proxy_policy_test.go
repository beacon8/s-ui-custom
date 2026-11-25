package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alireza0/s-ui/database/model"
)

func TestParseProxyProtocolOptions(t *testing.T) {
	options, trusted, err := parseProxyProtocolOptions(json.RawMessage(`{
		"enabled": true,
		"trustedCidrs": ["10.0.0.0/8", " 2001:db8::/32 ", ""],
		"headerTimeout": 3
	}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !options.Enabled || options.HeaderTimeout != 3 {
		t.Fatalf("unexpected options: %+v", options)
	}
	if len(trusted) != 2 || trusted[0].String() != "10.0.0.0/8" || trusted[1].String() != "2001:db8::/32" {
		t.Fatalf("unexpected trusted prefixes: %v", trusted)
	}
}

func TestParseProxyProtocolOptionsRejectsInvalidInput(t *testing.T) {
	cases := map[string]string{
		"empty trusted list": `{"enabled": true, "trustedCidrs": []}`,
		"invalid cidr":       `{"enabled": true, "trustedCidrs": ["not-a-cidr"]}`,
		"negative timeout":   `{"enabled": true, "trustedCidrs": ["10.0.0.0/8"], "headerTimeout": -1}`,
		"bad json":           `[]`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := parseProxyProtocolOptions(json.RawMessage(raw)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestGetProxyProtocolPolicies(t *testing.T) {
	db := clientTestDB(t)
	inbounds := []model.Inbound{
		{
			Type: "socks", Tag: "relay-in", Enable: true,
			Options:       json.RawMessage(`{"listen_port": 1080}`),
			ProxyProtocol: json.RawMessage(`{"enabled": true, "trustedCidrs": ["10.0.0.0/8"], "headerTimeout": 3}`),
		},
		{
			Type: "vless", Tag: "direct-in", Enable: true,
			Options:       json.RawMessage(`{"listen_port": 443}`),
			ProxyProtocol: json.RawMessage(`{"enabled": false}`),
		},
		{
			Type: "vmess", Tag: "disabled-in", Enable: false,
			Options:       json.RawMessage(`{"listen_port": 8443}`),
			ProxyProtocol: json.RawMessage(`{"enabled": true, "trustedCidrs": ["10.0.0.0/8"]}`),
		},
		{
			Type: "trojan", Tag: "invalid-in", Enable: true,
			Options:       json.RawMessage(`{"listen_port": 9443}`),
			ProxyProtocol: json.RawMessage(`{"enabled": true, "trustedCidrs": ["bad"]}`),
		},
	}
	for i := range inbounds {
		if err := db.Create(&inbounds[i]).Error; err != nil {
			t.Fatalf("create inbound: %v", err)
		}
	}
	// GORM reads the column default back for a false bool, so disable it
	// explicitly the way InboundService.Save does.
	if err := db.Model(&model.Inbound{}).Where("tag = ?", "disabled-in").Update("enable", false).Error; err != nil {
		t.Fatalf("disable inbound: %v", err)
	}
	service := &InboundService{}
	policies, err := service.GetProxyProtocolPolicies(db)
	if err != nil {
		t.Fatalf("policies: %v", err)
	}
	if len(policies) != 1 {
		t.Fatalf("policies = %+v, want only the valid relay inbound", policies)
	}
	policy, loaded := policies[1080]
	if !loaded {
		t.Fatalf("port 1080 missing: %+v", policies)
	}
	if !policy.Enabled || policy.InboundTag != "relay-in" || policy.HeaderTimeout != 3*time.Second {
		t.Fatalf("unexpected policy: %+v", policy)
	}
	if len(policy.Trusted) != 1 || policy.Trusted[0].String() != "10.0.0.0/8" {
		t.Fatalf("unexpected trusted prefixes: %v", policy.Trusted)
	}
}

func TestClientIPLimitMapAndDefaults(t *testing.T) {
	db := clientTestDB(t)
	createClient(t, db, &model.Client{Name: "limited", IPLimit: 2})
	createClient(t, db, &model.Client{Name: "unlimited", IPLimit: 0})
	service := &ClientService{}
	limits, err := service.GetIPLimitMap(db)
	if err != nil {
		t.Fatalf("GetIPLimitMap: %v", err)
	}
	if len(limits) != 1 || limits["limited"] != 2 {
		t.Fatalf("limits = %v", limits)
	}

	clients, err := service.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	found := false
	for _, client := range *clients {
		if client.Name == "limited" {
			found = true
			if client.IPLimit != 2 {
				t.Fatalf("ipLimit = %d, want 2", client.IPLimit)
			}
		}
	}
	if !found {
		t.Fatal("limited client missing from GetAll")
	}
}
