package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
)

func setupImportTest(t *testing.T) *model.Inbound {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(path); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		if db := database.GetDB(); db != nil {
			if sqlDB, err := db.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	})
	if _, err := (&SettingService{}).GetAllSetting(); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	inbound := &model.Inbound{
		Type:    "mixed",
		Tag:     "mixed-in",
		Enable:  true,
		Options: json.RawMessage(`{"listen":"127.0.0.1","listen_port":21080}`),
	}
	if err := database.GetDB().Create(inbound).Error; err != nil {
		t.Fatalf("create inbound: %v", err)
	}
	applyImportCore = nil
	return inbound
}

func importTestRequest(inboundID uint, lines ...string) *ImportRequest {
	nodes := make([]ImportNodeInput, 0, len(lines))
	for i := range lines {
		line := lines[i]
		nodes = append(nodes, ImportNodeInput{Raw: &line, Line: intPtr(i + 1)})
	}
	start, width := 1, 3
	enable := true
	volumeGiB := 100.0
	up, down := 10, 50
	return &ImportRequest{
		Prefix:          "US",
		StartID:         &start,
		IDWidth:         &width,
		InboundIDs:      []uint{inboundID},
		DefaultProtocol: "socks5",
		ClientDefaults: &ImportClientDefault{
			Enable:    &enable,
			VolumeGiB: &volumeGiB,
			UpMbps:    &up,
			DownMbps:  &down,
		},
		Nodes: nodes,
	}
}

func TestImportNodesCreatesNumberedBatch(t *testing.T) {
	inbound := setupImportTest(t)
	req := importTestRequest(inbound.Id,
		"192.0.2.10:1080:user1:p@ss:word",
		"192.0.2.11:1080:user2:pw2",
		"192.0.2.12:1080:user3:pw3",
	)
	res, ierr := (&ImportService{}).Import(req, "admin", "batch-1", "panel.example")
	if ierr != nil {
		t.Fatalf("import failed: %+v", ierr)
	}
	if res.BatchID == "" || res.Count != 3 || res.Group != "US-1-3" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if res.Core.Status != "skipped" {
		t.Errorf("core status = %q, want skipped without a running core", res.Core.Status)
	}
	wantNames := []string{"US-001", "US-002", "US-003"}
	wantTags := []string{"out-US-001", "out-US-002", "out-US-003"}
	for i, item := range res.Items {
		if item.ClientName != wantNames[i] || item.OutboundTag != wantTags[i] {
			t.Errorf("item %d = %s/%s, want %s/%s", i, item.ClientName, item.OutboundTag, wantNames[i], wantTags[i])
		}
		if item.Protocol != "socks5" {
			t.Errorf("item %d protocol = %q", i, item.Protocol)
		}
	}

	var clients []model.Client
	database.GetDB().Order("name").Find(&clients)
	if len(clients) != 3 {
		t.Fatalf("clients = %d, want 3", len(clients))
	}
	passwords := map[string]bool{}
	for _, client := range clients {
		if client.Group != "US-1-3" {
			t.Errorf("%s group = %q", client.Name, client.Group)
		}
		if client.Volume != 100*(1<<30) || client.UpMbps != 10 || client.DownMbps != 50 {
			t.Errorf("%s limits = volume %d up %d down %d", client.Name, client.Volume, client.UpMbps, client.DownMbps)
		}
		var configs map[string]map[string]any
		if err := json.Unmarshal(client.Config, &configs); err != nil {
			t.Fatal(err)
		}
		password, _ := configs["socks"]["password"].(string)
		passwords[password] = true
		if configs["socks"]["username"] != client.Name {
			t.Errorf("%s socks username = %v", client.Name, configs["socks"]["username"])
		}
		var inbounds []uint
		_ = json.Unmarshal(client.Inbounds, &inbounds)
		if len(inbounds) != 1 || inbounds[0] != inbound.Id {
			t.Errorf("%s inbounds = %v", client.Name, inbounds)
		}
	}
	if len(passwords) != 3 {
		t.Errorf("each client must get its own credentials, got %d distinct passwords", len(passwords))
	}

	var outbounds []model.Outbound
	database.GetDB().Where("tag LIKE ?", "out-US-%").Order("tag").Find(&outbounds)
	if len(outbounds) != 3 {
		t.Fatalf("outbounds = %d, want 3", len(outbounds))
	}
	if !strings.Contains(string(outbounds[0].Options), "p@ss:word") {
		t.Errorf("the raw password with ':' and '@' must survive: %s", outbounds[0].Options)
	}

	var configRow model.Setting
	database.GetDB().Where("key = ?", "config").First(&configRow)
	var base SingBoxConfig
	if err := json.Unmarshal([]byte(configRow.Value), &base); err != nil {
		t.Fatal(err)
	}
	var route map[string]json.RawMessage
	_ = json.Unmarshal(base.Route, &route)
	var rules []map[string]any
	_ = json.Unmarshal(route["rules"], &rules)
	if len(rules) != 2+3 {
		t.Fatalf("route rules = %d, want 5: %s", len(rules), route["rules"])
	}
	for i, name := range wantNames {
		rule := rules[2+i]
		users, _ := rule["auth_user"].([]any)
		if len(users) != 1 || users[0] != name || rule["outbound"] != wantTags[i] || rule["action"] != "route" {
			t.Errorf("rule %d = %v", i, rule)
		}
	}

	var changes []model.Changes
	database.GetDB().Where("key = ? AND action = ?", "clients", "import").Find(&changes)
	if len(changes) != 1 || changes[0].Actor != "admin" {
		t.Errorf("audit row = %+v", changes)
	}
	if strings.Contains(string(changes[0].Obj), "p@ss:word") {
		t.Error("audit must not contain credentials")
	}
}

func TestImportDryRunWritesNothing(t *testing.T) {
	inbound := setupImportTest(t)
	existing := &model.Client{Name: "US-001", Enable: true, Config: json.RawMessage(`{}`), Inbounds: json.RawMessage(`[]`), Links: json.RawMessage(`[]`)}
	if err := database.GetDB().Create(existing).Error; err != nil {
		t.Fatal(err)
	}
	var beforeClients, beforeOutbounds int64
	database.GetDB().Model(model.Client{}).Count(&beforeClients)
	database.GetDB().Model(model.Outbound{}).Count(&beforeOutbounds)
	req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p", "192.0.2.11:1080:u:p")
	req.DryRun = true
	res, ierr := (&ImportService{}).Import(req, "admin", "not-consumed", "")
	if ierr != nil {
		t.Fatalf("dry run failed: %+v", ierr)
	}
	if res.Valid || len(res.Errors) != 1 || res.Errors[0].Code != "CLIENT_NAME_CONFLICT" {
		t.Fatalf("expected one client name conflict, got %+v", res)
	}
	var afterClients, afterOutbounds int64
	database.GetDB().Model(model.Client{}).Count(&afterClients)
	database.GetDB().Model(model.Outbound{}).Count(&afterOutbounds)
	if afterClients != beforeClients || afterOutbounds != beforeOutbounds {
		t.Errorf("dry run wrote rows: clients %d->%d outbounds %d->%d", beforeClients, afterClients, beforeOutbounds, afterOutbounds)
	}
	var batchCount int64
	database.GetDB().Model(model.ImportBatch{}).Count(&batchCount)
	if batchCount != 0 {
		t.Errorf("dry run created %d import batch rows", batchCount)
	}

	// The dry-run key was not consumed: the real commit may use it. Start at 5
	// so the pre-created US-001 stays a conflict only for the preview.
	req.DryRun = false
	five := 5
	req.StartID = &five
	req.Nodes = req.Nodes[:1]
	commit, ierr := (&ImportService{}).Import(req, "admin", "not-consumed", "")
	if ierr != nil || commit.Valid == false {
		t.Fatalf("commit after dry run should work, got %+v / %+v", commit, ierr)
	}
}

func TestImportConflictStopsWholeBatch(t *testing.T) {
	inbound := setupImportTest(t)
	if err := database.GetDB().Create(&model.Outbound{Type: "socks", Tag: "out-US-002", Options: json.RawMessage(`{}`)}).Error; err != nil {
		t.Fatal(err)
	}
	req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p", "192.0.2.11:1080:u:p", "192.0.2.12:1080:u:p")
	_, ierr := (&ImportService{}).Import(req, "admin", "", "")
	if ierr == nil || ierr.Code != "VALIDATION_FAILED" {
		t.Fatalf("expected validation failure, got %+v", ierr)
	}
	if len(ierr.Errors) == 0 || ierr.Errors[0].Conflict != "out-US-002" {
		t.Errorf("expected the conflicting tag, got %+v", ierr.Errors)
	}
	var clients, outbounds int64
	database.GetDB().Model(model.Client{}).Count(&clients)
	database.GetDB().Model(model.Outbound{}).Where("tag LIKE ?", "out-US-%").Count(&outbounds)
	if clients != 0 || outbounds != 1 {
		t.Errorf("a failed batch wrote rows: clients=%d outbounds=%d", clients, outbounds)
	}
	var configRow model.Setting
	database.GetDB().Where("key = ?", "config").First(&configRow)
	if strings.Contains(configRow.Value, "out-US-001") {
		t.Error("a failed batch changed the stored route")
	}
}

func TestImportRejectsInvalidPortWithLine(t *testing.T) {
	inbound := setupImportTest(t)
	req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p", "192.0.2.11:70000:u:p")
	_, ierr := (&ImportService{}).Import(req, "admin", "", "")
	if ierr == nil || len(ierr.Errors) != 1 {
		t.Fatalf("expected one issue, got %+v", ierr)
	}
	if ierr.Errors[0].Line == nil || *ierr.Errors[0].Line != 2 {
		t.Errorf("line = %v, want 2", ierr.Errors[0].Line)
	}
}

func TestImportShareLinks(t *testing.T) {
	inbound := setupImportTest(t)
	uuid := "b831381d-6324-4d53-ad4f-8cda48b30811"
	vless := "vless://" + uuid + "@example.com:443?security=tls&sni=example.com&type=ws&path=%2Fws#vless-node"
	vmessPayload := map[string]any{
		"v": "2", "ps": "vmess-node", "add": "example.org", "port": "8443", "id": uuid,
		"aid": "0", "net": "ws", "type": "none", "host": "example.org", "path": "/ws",
		"tls": "tls", "sni": "example.org",
	}
	raw, _ := json.Marshal(vmessPayload)
	vmess := "vmess://" + base64.StdEncoding.EncodeToString(raw)
	https := "https://user:example-only@192.0.2.20:443#https-node"
	nodes := []ImportNodeInput{
		{URI: &vless, Line: intPtr(1)},
		{URI: &vmess, Line: intPtr(2)},
		{URI: &https, Line: intPtr(3)},
	}
	start, width := 1, 3
	req := &ImportRequest{Prefix: "MX", StartID: &start, IDWidth: &width, InboundIDs: []uint{inbound.Id}, Nodes: nodes}
	res, ierr := (&ImportService{}).Import(req, "admin", "", "")
	if ierr != nil {
		t.Fatalf("import share links: %+v errors=%+v", ierr, ierr.Errors)
	}
	want := map[string]string{"MX-001": "vless", "MX-002": "vmess", "MX-003": "https"}
	for _, item := range res.Items {
		if want[item.ClientName] != item.Protocol {
			t.Errorf("%s protocol = %q, want %q", item.ClientName, item.Protocol, want[item.ClientName])
		}
	}
	var outbounds []model.Outbound
	database.GetDB().Where("tag LIKE ?", "out-MX-%").Order("tag").Find(&outbounds)
	if len(outbounds) != 3 {
		t.Fatalf("outbounds = %d", len(outbounds))
	}
	if !strings.Contains(string(outbounds[0].Options), `"reality"`) && !strings.Contains(string(outbounds[0].Options), `"transport"`) {
		t.Errorf("vless transport/tls not preserved: %s", outbounds[0].Options)
	}
	if !strings.Contains(string(outbounds[1].Options), `"ws"`) {
		t.Errorf("vmess transport not preserved: %s", outbounds[1].Options)
	}
}

func TestImportIdempotency(t *testing.T) {
	inbound := setupImportTest(t)
	req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p")
	first, ierr := (&ImportService{}).Import(req, "admin", "same-key", "")
	if ierr != nil {
		t.Fatal(ierr)
	}
	second, ierr := (&ImportService{}).Import(req, "admin", "same-key", "")
	if ierr != nil {
		t.Fatal(ierr)
	}
	if !second.Replayed || second.BatchID != first.BatchID {
		t.Fatalf("replay = %+v, want batch %s", second, first.BatchID)
	}
	var count int64
	database.GetDB().Model(model.Client{}).Count(&count)
	if count != 1 {
		t.Errorf("replay created clients: %d", count)
	}

	changed := importTestRequest(inbound.Id, "192.0.2.11:1080:u:p")
	if _, ierr := (&ImportService{}).Import(changed, "admin", "same-key", ""); ierr == nil || ierr.Code != "IDEMPOTENCY_CONFLICT" {
		t.Fatalf("same key with another body = %+v, want IDEMPOTENCY_CONFLICT", ierr)
	}
}

func TestImportDifferentKeysRaceForSameRange(t *testing.T) {
	inbound := setupImportTest(t)
	type outcome struct {
		res *ImportResult
		err *ImportError
	}
	requests := 2
	results := make(chan outcome, requests)
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p", "192.0.2.11:1080:u:p")
			res, err := (&ImportService{}).Import(req, "admin", fmt.Sprintf("race-%d", n), "")
			results <- outcome{res, err}
		}(i)
	}
	wg.Wait()
	close(results)
	success := 0
	for out := range results {
		if out.err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successful batches = %d, want 1", success)
	}
	var clients int64
	database.GetDB().Model(model.Client{}).Count(&clients)
	if clients != 2 {
		t.Errorf("clients = %d, want 2 (one batch)", clients)
	}
}

func TestImportApplyFailureCanBeReapplied(t *testing.T) {
	inbound := setupImportTest(t)
	calls := 0
	applyImportCore = func() error {
		calls++
		if calls == 1 {
			return fmt.Errorf("injected core failure")
		}
		return nil
	}
	defer func() { applyImportCore = nil }()

	req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p")
	res, ierr := (&ImportService{}).Import(req, "admin", "", "")
	if res != nil {
		t.Fatal("a failed core application must not answer like a clean success")
	}
	if ierr == nil || ierr.Code != "CORE_APPLY_FAILED" || ierr.BatchID == "" {
		t.Fatalf("apply failure = %+v", ierr)
	}
	var clients int64
	database.GetDB().Model(model.Client{}).Count(&clients)
	if clients != 1 {
		t.Fatalf("clients after apply failure = %d, want 1 committed row", clients)
	}
	var batch model.ImportBatch
	if err := database.GetDB().Where("batch_id = ?", ierr.BatchID).First(&batch).Error; err != nil {
		t.Fatal(err)
	}
	if batch.Status != "apply_failed" {
		t.Errorf("batch status = %q", batch.Status)
	}

	applied, ierr := (&ImportService{}).Reapply(ierr.BatchID, "admin", "")
	if ierr != nil {
		t.Fatalf("reapply: %+v", ierr)
	}
	if applied.Core.Status != "applied" || applied.Count != 1 {
		t.Errorf("reapply result = %+v", applied)
	}
	database.GetDB().Model(model.Client{}).Count(&clients)
	if clients != 1 {
		t.Errorf("reapply must not recreate clients, got %d", clients)
	}
}

func TestImportRejectsInboundWithoutIdentity(t *testing.T) {
	_ = setupImportTest(t)
	inbound := &model.Inbound{Type: "tun", Tag: "tun-in", Enable: true, Options: json.RawMessage(`{}`)}
	if err := database.GetDB().Create(inbound).Error; err != nil {
		t.Fatal(err)
	}
	req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p")
	_, ierr := (&ImportService{}).Import(req, "admin", "", "")
	if ierr == nil || len(ierr.Errors) == 0 || ierr.Errors[0].Code != "INBOUND_NO_USER" {
		t.Fatalf("expected INBOUND_NO_USER, got %+v", ierr)
	}
}

func TestImportRouteOrderPreserved(t *testing.T) {
	inbound := setupImportTest(t)
	base := `{"route":{"rules":[
		{"action":"sniff"},
		{"protocol":["dns"],"action":"hijack-dns"},
		{"domain":["example.com"],"action":"route","outbound":"direct"},
		{"protocol":["quic"],"action":"reject"}
	]},"log":{"level":"info"}}`
	var row model.Setting
	database.GetDB().Where("key = ?", "config").First(&row)
	if err := database.GetDB().Model(&model.Setting{}).Where("key = ?", "config").Update("value", base).Error; err != nil {
		t.Fatal(err)
	}
	req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p")
	if _, ierr := (&ImportService{}).Import(req, "admin", "", ""); ierr != nil {
		t.Fatalf("import: %+v", ierr)
	}
	database.GetDB().Where("key = ?", "config").First(&row)
	var sb SingBoxConfig
	_ = json.Unmarshal([]byte(row.Value), &sb)
	var route map[string]json.RawMessage
	_ = json.Unmarshal(sb.Route, &route)
	var rules []map[string]any
	_ = json.Unmarshal(route["rules"], &rules)
	if len(rules) != 5 {
		t.Fatalf("rules = %d", len(rules))
	}
	if rules[0]["action"] != "sniff" || rules[1]["action"] != "hijack-dns" {
		t.Errorf("leading rules moved: %v", rules)
	}
	if rules[2]["outbound"] != "out-US-001" {
		t.Errorf("new rule not after pre-routing rules: %v", rules)
	}
	if rules[3]["domain"] == nil || rules[4]["action"] != "reject" {
		t.Errorf("existing rules changed: %v", rules)
	}
}

func TestImportRejectsSharedIdentityOverrides(t *testing.T) {
	inbound := setupImportTest(t)
	req := importTestRequest(inbound.Id, "192.0.2.10:1080:u:p")
	req.ClientDefaults.Config = map[string]map[string]any{
		"socks": {"password": "same-for-everyone"},
	}
	_, ierr := (&ImportService{}).Import(req, "admin", "", "")
	if ierr == nil {
		t.Fatal("a shared password must be rejected")
	}
	found := false
	for _, issue := range ierr.Errors {
		if issue.Code == "DEFAULTS_CONFIG_IDENTITY" {
			found = true
		}
	}
	if !found {
		t.Errorf("issues = %+v", ierr.Errors)
	}
}

func TestImportPreviewHidesCredentials(t *testing.T) {
	inbound := setupImportTest(t)
	req := importTestRequest(inbound.Id, "192.0.2.10:1080:demo-user:top-secret")
	req.DryRun = true
	res, ierr := (&ImportService{}).Import(req, "admin", "", "")
	if ierr != nil {
		t.Fatal(ierr)
	}
	encoded, _ := json.Marshal(res)
	if strings.Contains(string(encoded), "top-secret") {
		t.Error("preview leaked the outbound password")
	}
}
