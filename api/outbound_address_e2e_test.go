package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
)

// TestOutboundAddressesAPI covers the panel contract: the same GET action
// exists on the token router and returns the configured server address without
// requiring a running core or a network probe.
func TestOutboundAddressesAPI(t *testing.T) {
	srv, token, _ := apiv2TestServer(t)
	db := database.GetDB()

	baseConfig := `{
		"log": {"level": "error"},
		"route": {
			"final": "direct",
			"rules": [
				{"auth_user": ["alice"], "action": "route", "outbound": "socks-s28"},
				{"auth_user": ["bob"], "action": "route", "outbound": "direct"}
			]
		}
	}`
	if err := db.Where("key = ?", "config").Delete(&model.Setting{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Setting{Key: "config", Value: baseConfig}).Error; err != nil {
		t.Fatal(err)
	}
	for _, outbound := range []model.Outbound{
		{Type: "socks", Tag: "socks-s28", Options: json.RawMessage(`{"server":"1.1.1.1","server_port":80}`)},
		{Type: "direct", Tag: "direct"},
	} {
		outbound := outbound
		if err := db.Where("tag = ?", outbound.Tag).FirstOrCreate(&outbound).Error; err != nil {
			t.Fatal(err)
		}
	}

	requestURL := srv.URL + "/apiv2/outboundAddresses?users=" + url.QueryEscape(`["alice","bob"]`)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var msg Msg
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatal(err)
	}
	if !msg.Success {
		t.Fatalf("request failed: %s", msg.Msg)
	}
	raw, _ := json.Marshal(msg.Obj)
	var result struct {
		Users map[string]struct {
			Status   string `json:"status"`
			Outbound string `json:"outbound"`
			Address  string `json:"address"`
			Port     uint16 `json:"port"`
		} `json:"users"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if got := result.Users["alice"]; got.Status != "ok" || got.Outbound != "socks-s28" || got.Address != "1.1.1.1" || got.Port != 80 {
		t.Errorf("alice = %+v, want socks-s28 1.1.1.1:80: %s", got, raw)
	}
	if got := result.Users["bob"]; got.Status != "direct" {
		t.Errorf("bob = %+v, want direct: %s", got, raw)
	}

	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/apiv2/outboundAddresses?users=not-json", nil)
	req.Header.Set("Token", token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	msg = Msg{}
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatal(err)
	}
	if msg.Success {
		t.Fatal("invalid users parameter was accepted")
	}
}
