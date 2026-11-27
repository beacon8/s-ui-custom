package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
)

// TestReplaceOutboundsAPI covers the token route and the atomic batch update:
// the response includes the refreshed, flattened outbound list and keeps the
// original type/tag.
func TestReplaceOutboundsAPI(t *testing.T) {
	srv, token, _ := apiv2TestServer(t)
	db := database.GetDB()
	for _, row := range []model.Outbound{
		{Type: "socks", Tag: "socks-a", Options: json.RawMessage(`{"server":"192.0.2.1","server_port":1080,"version":"5"}`)},
		{Type: "http", Tag: "http-b", Options: json.RawMessage(`{"server":"192.0.2.2","server_port":8080,"tls":{"enabled":true}}`)},
	} {
		row := row
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	body := map[string]any{"items": []map[string]any{
		{"tag": "http-b", "server": "192.0.2.61", "server_port": 8080, "username": "user-b", "password": "pass:b"},
		{"tag": "socks-a", "server": "192.0.2.66", "server_port": 9873, "username": "user-a", "password": "pass-a"},
	}}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/apiv2/replaceOutbounds", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Token", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var msg struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		Obj     struct {
			Saved     int              `json:"saved"`
			Applied   bool             `json:"applied"`
			Warning   string           `json:"warning"`
			Outbounds []map[string]any `json:"outbounds"`
		} `json:"obj"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatal(err)
	}
	if !msg.Success || msg.Obj.Saved != 2 {
		t.Fatalf("response = %+v", msg)
	}
	if msg.Obj.Applied || msg.Obj.Warning == "" {
		t.Fatalf("core is not expected in this test: %+v", msg.Obj)
	}
	var stored model.Outbound
	if err := db.Where("tag = ?", "http-b").First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	var options map[string]json.RawMessage
	_ = json.Unmarshal(stored.Options, &options)
	var server string
	_ = json.Unmarshal(options["server"], &server)
	if server != "192.0.2.61" {
		t.Fatalf("server = %q", server)
	}
	if stored.Type != "http" || stored.Tag != "http-b" {
		t.Fatalf("identity changed: %+v", stored)
	}
}
