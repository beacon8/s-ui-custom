package sub

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/service"
)

func subscriptionTestDB(t *testing.T) *model.Inbound {
	t.Helper()
	if err := database.InitDB(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if db := database.GetDB(); db != nil {
			if sqlDB, err := db.DB(); err == nil {
				sqlDB.Close()
			}
		}
	})
	if _, err := (&service.SettingService{}).GetAllSetting(); err != nil {
		t.Fatal(err)
	}
	inbound := &model.Inbound{
		Type: "mixed", Tag: "sub-in", Enable: true,
		Options: json.RawMessage(`{"listen":"127.0.0.1","listen_port":23080}`),
	}
	if err := database.GetDB().Create(inbound).Error; err != nil {
		t.Fatal(err)
	}
	return inbound
}

func createSubscriptionFixture(t *testing.T) (*service.ImportResult, *model.Inbound) {
	t.Helper()
	inbound := subscriptionTestDB(t)
	start, width := 1, 3
	enable := true
	req := &service.ImportRequest{
		Prefix: "SUB", StartID: &start, IDWidth: &width, InboundIDs: []uint{inbound.Id},
		ClientDefaults: &service.ImportClientDefault{Enable: &enable},
		Nodes: []service.ImportNodeInput{
			{Raw: strPtr("192.0.2.10:1080:u:p")},
			{Raw: strPtr("192.0.2.11:1080:u:p")},
		},
	}
	result, err := (&service.ImportService{}).Import(req, "admin", "sub-key", "127.0.0.1")
	if err != nil {
		t.Fatalf("import: %+v", err)
	}
	return result, inbound
}

func strPtr(s string) *string { return &s }

func getSub(t *testing.T, srv *httptest.Server, path string) (int, string, http.Header) {
	t.Helper()
	resp, err := http.Get(srv.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), resp.Header
}

func TestGroupAndPersonalSubscriptionLinks(t *testing.T) {
	result, _ := createSubscriptionFixture(t)
	engine, err := NewServer().initRouter()
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(engine)
	defer srv.Close()
	subPath := "/sub/"

	if result.GroupSubscription.SubscriptionID == "" || len(result.Items) != 2 {
		t.Fatalf("missing subscription ids: %+v", result)
	}
	for _, item := range result.Items {
		if item.SubscriptionID == "" || !strings.Contains(item.SubscriptionLink, "/c/") {
			t.Errorf("item subscription = %+v", item)
		}
	}
	if !strings.Contains(result.GroupSubscription.Link, "/g/") {
		t.Errorf("group link = %q", result.GroupSubscription.Link)
	}
	var stored model.SubscriptionGroup
	if err := database.GetDB().Where("sub_id = ?", result.GroupSubscription.SubscriptionID).First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	groupToken := queryToken(t, result.GroupSubscription.Link)
	if strings.Contains(stored.TokenEnc, groupToken) {
		t.Error("group token is stored in the clear")
	}
	if decrypted, err := service.DecryptSubscriptionToken(stored.TokenEnc); err != nil || decrypted != groupToken {
		t.Errorf("decrypt group token = %q, %v", decrypted, err)
	}

	groupURL := strings.TrimPrefix(result.GroupSubscription.Link, "http://127.0.0.1")
	status, body, headers := getSub(t, srv, subPath+"g/"+result.GroupSubscription.SubscriptionID+"?token="+queryToken(t, result.GroupSubscription.Link))
	if status != 200 {
		t.Fatalf("group status = %d", status)
	}
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Split(strings.TrimSpace(string(decoded)), "\n"); len(lines) != 4 {
		t.Errorf("group lines = %d, want 4; decoded=%q", len(lines), decoded)
	}
	if headers.Get("Profile-Title") == "" {
		t.Error("group response has no title")
	}
	for _, format := range []string{"json", "clash"} {
		status, body, _ := getSub(t, srv, subPath+"g/"+result.GroupSubscription.SubscriptionID+"?format="+format+"&token="+groupToken)
		if status != 200 {
			t.Errorf("group %s = %d", format, status)
			continue
		}
		if format == "json" && !strings.Contains(body, `"outbounds"`) {
			t.Errorf("json group has no outbounds")
		}
		if format == "clash" && !strings.Contains(body, "proxies:") {
			t.Errorf("clash group has no proxies")
		}
	}
	_ = groupURL

	// Personal link returns one client only.
	personal := result.Items[0]
	pstatus, pbody, _ := getSub(t, srv, subPath+"c/"+personal.SubscriptionID+"?token="+queryToken(t, personal.SubscriptionLink))
	if pstatus != 200 {
		t.Fatalf("personal status = %d", pstatus)
	}
	pdecoded, _ := base64.StdEncoding.DecodeString(pbody)
	if lines := strings.Split(strings.TrimSpace(string(pdecoded)), "\n"); len(lines) != 2 {
		t.Errorf("personal lines = %d, want 2\n%s", len(lines), pdecoded)
	}
	if strings.Contains(string(pdecoded), result.Items[1].ClientName) {
		t.Error("personal response leaked the other member")
	}

	// Missing/wrong/cross credentials all fail.
	for _, path := range []string{
		subPath + "g/" + result.GroupSubscription.SubscriptionID,
		subPath + "g/" + result.GroupSubscription.SubscriptionID + "?token=wrong",
		subPath + "c/" + personal.SubscriptionID + "?token=" + queryToken(t, result.GroupSubscription.Link),
		subPath + "c/" + result.GroupSubscription.SubscriptionID + "?token=" + queryToken(t, personal.SubscriptionLink),
	} {
		if status, _, _ := getSub(t, srv, path); status != 404 {
			t.Errorf("%s = %d, want 404", path, status)
		}
	}

	// The old /<name> path is refused once a member row exists, including HEAD.
	if status, _, _ := getSub(t, srv, subPath+result.Items[0].ClientName); status != 400 {
		t.Errorf("legacy path = %d, want 400", status)
	}
	resp, err := http.Head(srv.URL + subPath + result.Items[0].ClientName)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Errorf("legacy HEAD = %d, want 400", resp.StatusCode)
	}

	// Disable the second member: group drops it, its personal link stops.
	database.GetDB().Model(&model.Client{}).Where("id = ?", result.Items[1].ClientID).Update("enable", false)
	status, body, _ = getSub(t, srv, subPath+"g/"+result.GroupSubscription.SubscriptionID+"?token="+queryToken(t, result.GroupSubscription.Link))
	if status != 200 {
		t.Fatal(status)
	}
	decoded, _ = base64.StdEncoding.DecodeString(body)
	if lines := strings.Split(strings.TrimSpace(string(decoded)), "\n"); len(lines) != 2 {
		t.Errorf("after disable group lines = %d, want 2", len(lines))
	}
	if status, _, _ := getSub(t, srv, subPath+"c/"+result.Items[1].SubscriptionID+"?token="+queryToken(t, result.Items[1].SubscriptionLink)); status != 404 {
		t.Errorf("disabled personal = %d, want 404", status)
	}

	// Rotating the first member's token invalidates only that personal link.
	oldPersonal := queryToken(t, result.Items[0].SubscriptionLink)
	var member model.SubscriptionMember
	if err := database.GetDB().Where("client_id = ?", result.Items[0].ClientID).First(&member).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.RotateMemberSubscriptionToken(member.Id); err != nil {
		t.Fatal(err)
	}
	if status, _, _ := getSub(t, srv, subPath+"c/"+result.Items[0].SubscriptionID+"?token="+oldPersonal); status != 404 {
		t.Errorf("rotated personal old token = %d, want 404", status)
	}
	// Group still works after a member rotation.
	if status, _, _ := getSub(t, srv, subPath+"g/"+result.GroupSubscription.SubscriptionID+"?token="+queryToken(t, result.GroupSubscription.Link)); status != 200 {
		t.Errorf("group after member rotation = %d", status)
	}

	// Rotating the group token invalidates the old group link but leaves
	// personal links alone.
	var groupRow model.SubscriptionGroup
	database.GetDB().Where("sub_id = ?", result.GroupSubscription.SubscriptionID).First(&groupRow)
	newToken, err := service.RotateGroupSubscriptionToken(groupRow.Id)
	if err != nil {
		t.Fatal(err)
	}
	if status, _, _ := getSub(t, srv, subPath+"g/"+groupRow.SubID+"?token="+groupToken); status != 404 {
		t.Errorf("old group token = %d, want 404", status)
	}
	if status, _, _ := getSub(t, srv, subPath+"g/"+groupRow.SubID+"?token="+newToken); status != 200 {
		t.Errorf("new group token = %d, want 200", status)
	}
	if status, _, _ := getSub(t, srv, subPath+"c/"+personal.SubscriptionID+"?token="+queryToken(t, personal.SubscriptionLink)); status != 200 {
		// personal token was rotated earlier, so build the current link
		var member model.SubscriptionMember
		database.GetDB().Where("sub_id = ?", personal.SubscriptionID).First(&member)
		current, _ := service.MemberSubscriptionLink("http://127.0.0.1:2096/sub/", &member)
		if status, _, _ := getSub(t, srv, subPath+"c/"+personal.SubscriptionID+"?token="+queryToken(t, current)); status != 200 {
			t.Errorf("personal after group rotation = %d", status)
		}
	}

	// Revoking the group leaves the personal link usable.
	service.SetGroupSubscriptionRevoked(groupRow.Id, true)
	if status, _, _ := getSub(t, srv, subPath+"g/"+groupRow.SubID+"?token="+newToken); status != 404 {
		t.Errorf("revoked group = %d, want 404", status)
	}
	if status, _, _ := getSub(t, srv, subPath+"c/"+personal.SubscriptionID+"?token="+queryToken(t, personal.SubscriptionLink)); status != 404 {
		t.Errorf("revoked group must not revoke personal link")
	}
}

func queryToken(t *testing.T, link string) string {
	t.Helper()
	if i := strings.Index(link, "token="); i >= 0 {
		return link[i+len("token="):]
	}
	t.Fatalf("no token in %q", link)
	return ""
}
