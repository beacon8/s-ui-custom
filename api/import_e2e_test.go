package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
)

func importInbound(t *testing.T) uint {
	t.Helper()
	inbound := &model.Inbound{
		Type:    "mixed",
		Tag:     "mixed-e2e",
		Enable:  true,
		Options: json.RawMessage(`{"listen":"127.0.0.1","listen_port":21081}`),
	}
	if err := database.GetDB().Create(inbound).Error; err != nil {
		t.Fatal(err)
	}
	return inbound.Id
}

func postImportJSON(t *testing.T, srvURL, token, idempotencyKey string, body any) (int, Msg) {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, srvURL+"/apiv2/importClientNodes", bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Token", token)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var msg Msg
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, msg
}

func TestImportClientNodesAPI(t *testing.T) {
	srv, token, _ := apiv2TestServer(t)
	inboundID := importInbound(t)

	base := map[string]any{
		"prefix":     "API",
		"startId":    1,
		"idWidth":    3,
		"inboundIds": []uint{inboundID},
		"clientDefaults": map[string]any{
			"enable": true, "volumeGiB": 100, "maxConns": 8, "ipLimit": 2, "upMbps": 10, "downMbps": 50,
		},
		"nodes": []map[string]any{
			{"raw": "192.0.2.10:1080:demo-user:example-only", "line": 1},
			{"raw": "192.0.2.11:8080:demo-user:example-only", "line": 2},
		},
	}

	dry := map[string]any{}
	for k, v := range base {
		dry[k] = v
	}
	dry["dryRun"] = true
	status, msg := postImportJSON(t, srv.URL, token, "not-used-by-dry-run", dry)
	if status != http.StatusOK || !msg.Success {
		t.Fatalf("dry run = %d %+v", status, msg)
	}
	var dryResult map[string]any
	raw, _ := json.Marshal(msg.Obj)
	_ = json.Unmarshal(raw, &dryResult)
	if dryResult["valid"] != true || dryResult["count"].(float64) != 2 || dryResult["group"] != "API-1-2" {
		t.Fatalf("dry result = %s", raw)
	}
	var clientsAfterDry int64
	database.GetDB().Model(model.Client{}).Count(&clientsAfterDry)

	status, msg = postImportJSON(t, srv.URL, token, "api-e2e-key", base)
	if status != http.StatusOK || !msg.Success {
		t.Fatalf("commit = %d %+v", status, msg)
	}
	var result map[string]any
	raw, _ = json.Marshal(msg.Obj)
	_ = json.Unmarshal(raw, &result)
	if result["batchId"] == "" || result["count"].(float64) != 2 {
		t.Fatalf("result = %s", raw)
	}
	groupSub, _ := result["groupSubscription"].(map[string]any)
	if groupSub == nil || groupSub["subscriptionId"] == "" || !strings.Contains(groupSub["link"].(string), "/g/") {
		t.Fatalf("group subscription = %s", raw)
	}
	items, _ := result["items"].([]any)
	firstItem, _ := items[0].(map[string]any)
	if firstItem["subscriptionId"] == "" || !strings.Contains(firstItem["subscriptionLink"].(string), "/c/") {
		t.Fatalf("personal subscription = %s", raw)
	}
	batchID, _ := result["batchId"].(string)
	if strings.Contains(string(raw), "example-only") {
		t.Errorf("response leaked the upstream password: %s", raw)
	}
	var clients int64
	database.GetDB().Model(model.Client{}).Count(&clients)
	if clients != clientsAfterDry+2 {
		t.Errorf("clients = %d, dry run left %d", clients, clientsAfterDry)
	}

	// Same key, same body: the stored batch, not a second one.
	status, msg = postImportJSON(t, srv.URL, token, "api-e2e-key", base)
	if status != http.StatusOK || !msg.Success {
		t.Fatalf("replay = %d %+v", status, msg)
	}
	var replay map[string]any
	raw, _ = json.Marshal(msg.Obj)
	_ = json.Unmarshal(raw, &replay)
	if replay["batchId"] != batchID || replay["replayed"] != true {
		t.Errorf("replay = %s", raw)
	}
	replayGroup, _ := replay["groupSubscription"].(map[string]any)
	if replayGroup["subscriptionId"] != groupSub["subscriptionId"] {
		t.Errorf("replay group id changed: %v -> %v", groupSub["subscriptionId"], replayGroup["subscriptionId"])
	}

	// Same key, different body: a conflict, not a silent overwrite.
	changed := map[string]any{}
	for k, v := range base {
		changed[k] = v
	}
	changed["nodes"] = []map[string]any{{"raw": "192.0.2.12:1080:u:p", "line": 1}}
	status, msg = postImportJSON(t, srv.URL, token, "api-e2e-key", changed)
	if status != http.StatusConflict || msg.Success {
		t.Fatalf("idempotency conflict = %d %+v", status, msg)
	}

	// Bad input is rejected with a stable code, line number and no writes.
	bad := map[string]any{}
	for k, v := range base {
		bad[k] = v
	}
	bad["nodes"] = []map[string]any{
		{"raw": "192.0.2.20:1080:u:p", "line": 1},
		{"raw": "192.0.2.21:70000:u:p", "line": 2},
	}
	status, msg = postImportJSON(t, srv.URL, token, "", bad)
	if status != http.StatusUnprocessableEntity || msg.Success {
		t.Fatalf("bad batch = %d %+v", status, msg)
	}
	detail, _ := msg.Obj.(map[string]any)
	if detail["code"] != "VALIDATION_FAILED" {
		t.Errorf("error code = %v", detail["code"])
	}

	// Token is required and is not replaced by a cookie.
	status, msg = postImportJSON(t, srv.URL, "", "", base)
	if status != http.StatusUnauthorized || msg.Success {
		t.Fatalf("no token = %d %+v", status, msg)
	}

	// The batch can be located later and reapplied without creating anything.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/apiv2/importBatch?batchId="+batchID, nil)
	req.Header.Set("Token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var viewMsg Msg
	_ = json.NewDecoder(resp.Body).Decode(&viewMsg)
	if resp.StatusCode != http.StatusOK || !viewMsg.Success {
		t.Fatalf("importBatch = %d %+v", resp.StatusCode, viewMsg)
	}
	viewRaw, _ := json.Marshal(viewMsg)
	if strings.Contains(string(viewRaw), "example-only") {
		t.Errorf("importBatch leaked the upstream password: %s", viewRaw)
	}
	if strings.Contains(string(viewRaw), "token=") {
		t.Errorf("importBatch leaked a subscription token: %s", viewRaw)
	}

	// Admin management: list groups, open one, and fetch the personal link
	// for an already imported client. Tokens appear only on these explicit
	// authorised reads, never in the batch view.
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/apiv2/subscriptionGroups", nil)
	req.Header.Set("Token", token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var groupsMsg Msg
	_ = json.NewDecoder(resp.Body).Decode(&groupsMsg)
	resp.Body.Close()
	groupsRaw, _ := json.Marshal(groupsMsg.Obj)
	if !groupsMsg.Success || strings.Contains(string(groupsRaw), "token=") {
		t.Fatalf("subscriptionGroups = %s", groupsRaw)
	}
	var groupRows []map[string]any
	_ = json.Unmarshal(groupsRaw, &groupRows)
	if len(groupRows) == 0 {
		t.Fatal("no subscription groups")
	}
	groupID := int(groupRows[0]["id"].(float64))
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/apiv2/subscriptionGroup?groupId=%d", srv.URL, groupID), nil)
	req.Header.Set("Token", token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var groupMsg Msg
	_ = json.NewDecoder(resp.Body).Decode(&groupMsg)
	resp.Body.Close()
	groupDetail, _ := json.Marshal(groupMsg.Obj)
	if !groupMsg.Success || !strings.Contains(string(groupDetail), "/g/") {
		t.Fatalf("subscriptionGroup = %s", groupDetail)
	}

	enableForm := url.Values{"clientId": {strconv.FormatUint(uint64(firstItem["clientId"].(float64)), 10)}}
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/apiv2/enableClientSubscription", strings.NewReader(enableForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Token", token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var enableMsg Msg
	_ = json.NewDecoder(resp.Body).Decode(&enableMsg)
	resp.Body.Close()
	enableRaw, _ := json.Marshal(enableMsg.Obj)
	if !enableMsg.Success || !strings.Contains(string(enableRaw), "/c/") {
		t.Fatalf("enableClientSubscription = %s", enableRaw)
	}
	if !strings.Contains(string(enableRaw), firstItem["subscriptionId"].(string)) {
		t.Errorf("enableClientSubscription changed the personal id: %s", enableRaw)
	}

	reapplyReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/apiv2/reapplyImportBatch", strings.NewReader(`{"batchId":"`+batchID+`"}`))
	reapplyReq.Header.Set("Content-Type", "application/json")
	reapplyReq.Header.Set("Token", token)
	reapplyResp, err := http.DefaultClient.Do(reapplyReq)
	if err != nil {
		t.Fatal(err)
	}
	defer reapplyResp.Body.Close()
	var reapplyMsg Msg
	_ = json.NewDecoder(reapplyResp.Body).Decode(&reapplyMsg)
	if reapplyResp.StatusCode != http.StatusOK || !reapplyMsg.Success {
		t.Fatalf("reapply = %d %+v", reapplyResp.StatusCode, reapplyMsg)
	}
	database.GetDB().Model(model.Client{}).Count(&clients)
	if clients != 2 {
		t.Errorf("reapply recreated clients: %d", clients)
	}
}
