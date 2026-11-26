package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
)

func postTokenForm(t *testing.T, srv *httptestServer, token, action string, values url.Values) Msg {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/apiv2/"+action, strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body Msg
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func getTokenAPI(t *testing.T, srv *httptestServer, token, path string) (int, Msg) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/apiv2/"+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Token", token)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body Msg
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body
}

// httptestServer is the subset used here, so the helpers do not depend on the
// full httptest package in signatures.
type httptestServer = struct {
	URL    string
	Client func() *http.Client
}

func TestTokenLifecycleAndOwnership(t *testing.T) {
	srv, adminToken, handler := apiv2TestServer(t)
	server := &httptestServer{URL: srv.URL, Client: srv.Client}
	db := database.GetDB()

	bob := &model.User{Username: "bob", Password: "bob"}
	if err := db.Create(bob).Error; err != nil {
		t.Fatal(err)
	}
	bobTokenValue := "bob-token-32-characters-long!!!!"
	bobToken := &model.Tokens{Token: bobTokenValue, Desc: "bob", UserId: bob.Id}
	if err := db.Create(bobToken).Error; err != nil {
		t.Fatal(err)
	}
	expired := &model.Tokens{Token: "expired-token-32-characters!!!", Desc: "expired", Expiry: time.Now().Unix() - 10, UserId: bob.Id}
	if err := db.Create(expired).Error; err != nil {
		t.Fatal(err)
	}
	handler.ReloadTokens()

	// Expired tokens fail on the very next request.
	if status, body := getTokenAPI(t, server, expired.Token, "status"); status != http.StatusUnauthorized || body.Success {
		t.Errorf("expired token = %d %+v, want 401", status, body)
	}
	// A valid token owned by another admin works for that admin.
	if status, body := getTokenAPI(t, server, bobTokenValue, "status"); status != http.StatusOK || !body.Success {
		t.Errorf("bob token = %d %+v, want 200", status, body)
	}

	// The token list is the caller's own list and is masked.
	_, list := getTokenAPI(t, server, adminToken, "tokens")
	encoded, _ := json.Marshal(list.Obj)
	if !strings.Contains(string(encoded), `"****"`) {
		t.Errorf("token list is not masked: %s", encoded)
	}
	if strings.Contains(string(encoded), bobTokenValue) || strings.Contains(string(encoded), adminToken) {
		t.Errorf("token list leaked a token: %s", encoded)
	}
	var listed []map[string]any
	_ = json.Unmarshal(encoded, &listed)
	if len(listed) != 1 {
		t.Errorf("admin sees %d tokens, want only its own 1", len(listed))
	}

	// Deleting another admin's token is refused and leaves it usable.
	del := postTokenForm(t, server, adminToken, "deleteToken", url.Values{"id": {strconv.FormatUint(uint64(bobToken.Id), 10)}})
	if del.Success {
		t.Error("admin deleted another admin's token")
	}
	if status, body := getTokenAPI(t, server, bobTokenValue, "status"); status != http.StatusOK || !body.Success {
		t.Errorf("bob token was revoked by another admin: %d %+v", status, body)
	}

	// A token created over the token API is owned by the caller and works.
	added := postTokenForm(t, server, adminToken, "addToken", url.Values{"desc": {"created-via-token"}, "expiry": {"1"}})
	if !added.Success {
		t.Fatalf("addToken: %+v", added)
	}
	newToken, _ := added.Obj.(string)
	if len(newToken) != 32 {
		t.Fatalf("new token = %q", newToken)
	}
	if status, body := getTokenAPI(t, server, newToken, "status"); status != http.StatusOK || !body.Success {
		t.Errorf("new token does not work: %d %+v", status, body)
	}
	var newRow model.Tokens
	if err := db.Where("token = ?", newToken).First(&newRow).Error; err != nil {
		t.Fatal(err)
	}
	var admin model.User
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if newRow.UserId != admin.Id {
		t.Errorf("new token owner = %d, want admin %d", newRow.UserId, admin.Id)
	}
	if del := postTokenForm(t, server, adminToken, "deleteToken", url.Values{"id": {strconv.FormatUint(uint64(newRow.Id), 10)}}); !del.Success {
		t.Fatalf("revoke: %+v", del)
	}
	// Revocation is effective on the next request.
	if status, body := getTokenAPI(t, server, newToken, "status"); status != http.StatusUnauthorized || body.Success {
		t.Errorf("revoked token = %d %+v, want 401", status, body)
	}

	// changePass uses the token owner's identity and keeps the token usable.
	changed := postTokenForm(t, server, adminToken, "changePass", url.Values{
		"oldPass": {"admin"}, "newUsername": {"admin"}, "newPass": {"changed-pass"},
	})
	if !changed.Success {
		t.Fatalf("changePass: %+v", changed)
	}
	if status, body := getTokenAPI(t, server, adminToken, "status"); status != http.StatusOK || !body.Success {
		t.Errorf("token after changePass = %d %+v", status, body)
	}
	reverted := postTokenForm(t, server, adminToken, "changePass", url.Values{
		"oldPass": {"changed-pass"}, "newUsername": {"admin"}, "newPass": {"admin"},
	})
	if !reverted.Success {
		t.Fatalf("changePass revert: %+v", reverted)
	}
}
