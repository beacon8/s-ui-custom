package service

import (
	"encoding/json"
	"testing"

	"github.com/alireza0/s-ui/database/model"
)

func TestGetRateLimitsMapIncludesRateOnlyClients(t *testing.T) {
	db := clientTestDB(t)
	createClient(t, db, &model.Client{Name: "rate-only", UpMbps: 2, DownMbps: 8})
	createClient(t, db, &model.Client{Name: "conns-only", MaxConns: 5})
	createClient(t, db, &model.Client{Name: "unlimited"})

	got, err := (&ClientService{}).GetRateLimitsMap(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected one rate-limited client, got %v", got)
	}
	rate, ok := got["rate-only"]
	if !ok || rate.UpMbps != 2 || rate.DownMbps != 8 {
		t.Fatalf("unexpected rate map: %v", got)
	}
}

func TestPreserveMissingRates(t *testing.T) {
	db := clientTestDB(t)
	c := createClient(t, db, &model.Client{Name: "keep-rates", UpMbps: 5, DownMbps: 9})
	s := &ClientService{}

	// Both keys absent: keep both stored values.
	client := model.Client{Id: c.Id}
	s.preserveMissingRates(db, &client, json.RawMessage(`{"name":"keep-rates"}`))
	if client.UpMbps != 5 || client.DownMbps != 9 {
		t.Fatalf("missing fields were not preserved: up=%d down=%d", client.UpMbps, client.DownMbps)
	}

	// Explicit zero clears only the field that was sent.
	client = model.Client{Id: c.Id, UpMbps: 0, DownMbps: 0}
	s.preserveMissingRates(db, &client, json.RawMessage(`{"upMbps":0}`))
	if client.UpMbps != 0 || client.DownMbps != 9 {
		t.Fatalf("explicit zero did not win: up=%d down=%d", client.UpMbps, client.DownMbps)
	}
}

func TestPreserveMissingRatesBulk(t *testing.T) {
	db := clientTestDB(t)
	a := createClient(t, db, &model.Client{Name: "bulk-a", UpMbps: 1, DownMbps: 2})
	b := createClient(t, db, &model.Client{Name: "bulk-b", UpMbps: 3, DownMbps: 4})
	s := &ClientService{}

	clients := []*model.Client{
		{Id: a.Id, Name: "bulk-a"},
		{Id: b.Id, Name: "bulk-b", UpMbps: 0, DownMbps: 0},
	}
	s.preserveMissingRatesBulk(db, clients, json.RawMessage(`[{"id":1,"name":"bulk-a"},{"id":2,"name":"bulk-b","upMbps":0,"downMbps":0}]`))
	if clients[0].UpMbps != 1 || clients[0].DownMbps != 2 {
		t.Fatalf("missing bulk fields were not preserved: %+v", clients[0])
	}
	if clients[1].UpMbps != 0 || clients[1].DownMbps != 0 {
		t.Fatalf("explicit bulk zeros were not kept: %+v", clients[1])
	}
}

func TestValidateClientRatesBoundary(t *testing.T) {
	client := &model.Client{UpMbps: 17179, DownMbps: 17179}
	if err := validateClientRates(client); err != nil {
		t.Fatalf("17179 Mbps rejected: %v", err)
	}
	client.UpMbps = 17180
	if err := validateClientRates(client); err == nil {
		t.Fatal("17180 Mbps upstream must be rejected")
	}
	client.UpMbps = 0
	client.DownMbps = 17180
	if err := validateClientRates(client); err == nil {
		t.Fatal("17180 Mbps downstream must be rejected")
	}
	client = &model.Client{UpMbps: -1, DownMbps: -1}
	if err := validateClientRates(client); err != nil {
		t.Fatalf("negative rates should clamp to zero: %v", err)
	}
	if client.UpMbps != 0 || client.DownMbps != 0 {
		t.Fatalf("negative rates were not clamped: %+v", client)
	}
}

func TestGetRateLimitsMapRejectsOverflow(t *testing.T) {
	db := clientTestDB(t)
	createClient(t, db, &model.Client{Name: "too-fast", UpMbps: 17180})
	if _, err := (&ClientService{}).GetRateLimitsMap(db); err == nil {
		t.Fatal("expected an out-of-range error for 17180 Mbps")
	}
}

func TestSaveRejectsOverflowRate(t *testing.T) {
	db := clientTestDB(t)
	data := json.RawMessage(`{"name":"too-fast","enable":true,"inbounds":[],"links":[],"config":{},"upMbps":17180}`)
	if _, err := (&ClientService{}).Save(db, "new", data, ""); err == nil {
		t.Fatal("Save accepted 17180 Mbps")
	}
}
