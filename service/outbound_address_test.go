package service

import (
	"encoding/json"
	"testing"
)

func egressConfig(t *testing.T, body string) json.RawMessage {
	t.Helper()
	if !json.Valid([]byte(body)) {
		t.Fatalf("invalid test config: %s", body)
	}
	return json.RawMessage(body)
}

func TestResolveEgressTargetsRouteRules(t *testing.T) {
	cases := []struct {
		name   string
		config string
		users  []string
		want   map[string]egressTarget
	}{
		{
			name: "imported single-user rules beat the final outbound",
			config: `{
				"outbounds": [{"tag": "out-alice"}, {"tag": "out-bob"}, {"tag": "shared"}, {"tag": "direct"}],
				"route": {"final": "direct", "rules": [
					{"action": "sniff"},
					{"action": "hijack-dns", "protocol": ["dns"]},
					{"auth_user": ["alice"], "action": "route", "outbound": "out-alice"},
					{"auth_user": ["bob"], "action": "route", "outbound": "shared"},
					{"auth_user": ["carol"], "action": "route", "outbound": "shared"},
					{"domain_suffix": ["example.com"], "action": "route", "outbound": "out-bob"}
				]}
			}`,
			users: []string{"alice", "bob", "carol"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetOutbound, tag: "out-alice"},
				"bob":   {kind: egressTargetOutbound, tag: "shared"},
				"carol": {kind: egressTargetOutbound, tag: "shared"},
			},
		},
		{
			name: "earlier domain rule makes a user split",
			config: `{
				"outbounds": [{"tag": "out-alice"}, {"tag": "out-bob"}, {"tag": "direct"}],
				"route": {"final": "direct", "rules": [
					{"action": "sniff"},
					{"domain_suffix": ["example.com"], "action": "route", "outbound": "out-bob"},
					{"auth_user": ["alice"], "action": "route", "outbound": "out-alice"}
				]}
			}`,
			users: []string{"alice"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetSplit},
			},
		},
		{
			name: "rules for other users do not disturb the final outbound",
			config: `{
				"outbounds": [{"tag": "out-bob"}, {"tag": "direct"}],
				"route": {"final": "direct", "rules": [
					{"auth_user": ["bob"], "action": "route", "outbound": "out-bob"}
				]}
			}`,
			users: []string{"dave"},
			want: map[string]egressTarget{
				"dave": {kind: egressTargetOutbound, tag: "direct"},
			},
		},
		{
			name: "logical and with a domain is split",
			config: `{
				"outbounds": [{"tag": "out-alice"}, {"tag": "direct"}],
				"route": {"final": "direct", "rules": [
					{"type": "logical", "mode": "and", "rules": [
						{"auth_user": ["alice"]},
						{"domain": ["example.org"]}
					], "action": "route", "outbound": "out-alice"}
				]}
			}`,
			users: []string{"alice"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetSplit},
			},
		},
		{
			name: "logical or that is unconditionally true for a user is final",
			config: `{
				"outbounds": [{"tag": "shared"}, {"tag": "direct"}],
				"route": {"final": "direct", "rules": [
					{"type": "logical", "mode": "or", "rules": [
						{"auth_user": ["alice"]},
						{"auth_user": ["bob"]}
					], "action": "route", "outbound": "shared"}
				]}
			}`,
			users: []string{"alice"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetOutbound, tag: "shared"},
			},
		},
		{
			name: "inverted auth_user excludes only that user",
			config: `{
				"outbounds": [{"tag": "out-bob"}, {"tag": "direct"}],
				"route": {"final": "direct", "rules": [
					{"auth_user": ["bob"], "invert": true, "action": "route", "outbound": "out-bob"}
				]}
			}`,
			users: []string{"alice", "bob"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetOutbound, tag: "out-bob"},
				"bob":   {kind: egressTargetOutbound, tag: "direct"},
			},
		},
		{
			name: "reject is not an outbound",
			config: `{
				"outbounds": [{"tag": "direct"}],
				"route": {"final": "direct", "rules": [
					{"auth_user": ["alice"], "action": "reject"}
				]}
			}`,
			users: []string{"alice"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetUnbound},
			},
		},
		{
			name: "missing outbound tag is unbound",
			config: `{
				"outbounds": [{"tag": "direct"}],
				"route": {"final": "direct", "rules": [
					{"auth_user": ["alice"], "action": "route", "outbound": "deleted"}
				]}
			}`,
			users: []string{"alice"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetUnbound},
			},
		},
		{
			name: "string auth_user and missing final use the first outbound",
			config: `{
				"outbounds": [{"tag": "first"}, {"tag": "second"}],
				"route": {"rules": [{"auth_user": "alice", "action": "route", "outbound": "second"}]}
			}`,
			users: []string{"alice", "bob"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetOutbound, tag: "second"},
				"bob":   {kind: egressTargetOutbound, tag: "first"},
			},
		},
		{
			name:   "no rules and no outbounds is unbound",
			config: `{"route": {}}`,
			users:  []string{"alice"},
			want: map[string]egressTarget{
				"alice": {kind: egressTargetUnbound},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveEgressTargets(egressConfig(t, tc.config), tc.users)
			if err != nil {
				t.Fatal(err)
			}
			for user, want := range tc.want {
				if got[user] != want {
					t.Fatalf("%s = %+v, want %+v (all: %+v)", user, got[user], want, got)
				}
			}
		})
	}
}

func TestOutboundAddressesFromConfig(t *testing.T) {
	config := `{
		"outbounds": [
			{"tag": "socks-s28", "type": "socks", "server": "1.1.1.1", "server_port": 80},
			{"tag": "direct", "type": "direct"},
			{"tag": "auto", "type": "selector", "outbounds": ["socks-s28", "direct"]},
			{"tag": "blocked", "type": "block"},
			{"tag": "wg", "type": "wireguard"}
		],
		"route": {"final": "direct", "rules": [
			{"auth_user": ["alice"], "action": "route", "outbound": "socks-s28"},
			{"auth_user": ["bob"], "action": "route", "outbound": "direct"},
			{"auth_user": ["carol"], "action": "route", "outbound": "auto"},
			{"auth_user": ["dave"], "action": "route", "outbound": "blocked"},
			{"auth_user": ["erin"], "action": "route", "outbound": "wg"},
			{"domain_suffix": ["split.example"], "action": "route", "outbound": "socks-s28"}
		]}
	}`
	var parsed egressRouteConfig
	if err := json.Unmarshal([]byte(config), &parsed); err != nil {
		t.Fatal(err)
	}
	got := resolveOutboundAddresses(parsed, []string{"alice", "bob", "carol", "dave", "erin", "frank"})
	cases := []struct {
		user string
		want OutboundAddress
	}{
		{"alice", OutboundAddress{Status: OutboundAddressStatusOK, Outbound: "socks-s28", Address: "1.1.1.1", Port: 80, Type: "socks"}},
		{"bob", OutboundAddress{Status: OutboundAddressStatusDirect, Outbound: "direct", Type: "direct"}},
		{"carol", OutboundAddress{Status: OutboundAddressStatusMulti, Outbound: "auto", Type: "selector"}},
		{"dave", OutboundAddress{Status: OutboundAddressStatusUnbound}},
		{"erin", OutboundAddress{Status: OutboundAddressStatusNone, Outbound: "wg", Type: "wireguard"}},
		{"frank", OutboundAddress{Status: OutboundAddressStatusSplit}},
	}
	for _, tc := range cases {
		if got[tc.user] != tc.want {
			t.Fatalf("%s = %+v, want %+v", tc.user, got[tc.user], tc.want)
		}
	}
}

func TestOutboundAddressesMissingTagIsUnbound(t *testing.T) {
	var parsed egressRouteConfig
	config := `{
		"outbounds": [{"tag": "socks-s28", "type": "socks", "server": "1.1.1.1", "server_port": 80}],
		"route": {"final": "socks-s28", "rules": [
			{"auth_user": ["alice"], "action": "route", "outbound": "deleted"}
		]}
	}`
	if err := json.Unmarshal([]byte(config), &parsed); err != nil {
		t.Fatal(err)
	}
	got := resolveOutboundAddresses(parsed, []string{"alice"})
	if got["alice"].Status != OutboundAddressStatusUnbound {
		t.Fatalf("alice = %+v, want unbound", got["alice"])
	}
}
