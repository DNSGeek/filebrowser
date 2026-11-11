package fbhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/DNSGeek/filebrowser/v3/settings"
	"github.com/DNSGeek/filebrowser/v3/users"
)

// Hooks run commands on the host, so an administrator of the web interface
// must be able neither to read nor to change them: a stolen admin session
// would otherwise be code execution on the server.
func TestSettingsAPIDoesNotExposeHooks(t *testing.T) {
	key := []byte("test-signing-key")
	perm := users.Permissions{Admin: true}
	st := scopedUserStorage(t, t.TempDir(), perm, key)

	hooks := map[string][]string{"after_upload": {"/usr/local/bin/scan $FILE"}}
	if err := st.Settings.Save(&settings.Settings{Key: key, Hooks: hooks}); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Auth", signToken(t, perm, key))
	rec := httptest.NewRecorder()
	handle(settingsGetHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET settings: %d %s", rec.Code, rec.Body.String())
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"hooks", "commands"} {
		if _, ok := got[field]; ok {
			t.Errorf("GET settings exposes %q", field)
		}
	}

	body := `{"hooks":{"after_upload":["touch /tmp/pwned"]},"commands":{"after_upload":["touch /tmp/pwned"]},"minimumPasswordLength":12}`
	req, _ = http.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set("X-Auth", signToken(t, perm, key))
	rec = httptest.NewRecorder()
	handle(settingsPutHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT settings: %d %s", rec.Code, rec.Body.String())
	}

	saved, err := st.Settings.Get()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(saved.Hooks["after_upload"], hooks["after_upload"]) {
		t.Errorf("hooks changed over HTTP: %v", saved.Hooks["after_upload"])
	}
	if saved.HasLegacyCommands() {
		t.Errorf("legacy commands written over HTTP: %v", saved.Commands)
	}
}
