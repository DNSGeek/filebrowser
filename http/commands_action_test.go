package fbhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/DNSGeek/filebrowser/v3/files"
	"github.com/DNSGeek/filebrowser/v3/rules"
	"github.com/DNSGeek/filebrowser/v3/settings"
	"github.com/DNSGeek/filebrowser/v3/users"
	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
)

// commandData builds the request data of a user scoped to <root>/user who may
// run the "ls" and "pull" actions. <root>/outside holds a file the user must
// never reach.
func commandData(t *testing.T) (d *data, scope, outside string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX paths")
	}

	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	scope, outside = filepath.Join(root, "user"), filepath.Join(root, "outside")
	for _, dir := range []string{filepath.Join(scope, "docs"), filepath.Join(scope, "private"), outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("s"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret"), filepath.Join(scope, "docs", "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(scope, "dirlink")); err != nil {
		t.Fatal(err)
	}

	d = &data{
		settings: &settings.Settings{Actions: map[string]string{
			"ls":      "/bin/ls -la -- $FILES",
			"pull":    "/usr/bin/git pull",
			"secret":  "/bin/cat " + filepath.Join(outside, "secret"),
			"relpath": "ls $FILES",
		}},
		server: &settings.Server{EnableExec: true},
		user: &users.User{
			Username: "u",
			Perm:     users.Permissions{Execute: true},
			Commands: []string{"ls", "pull", "relpath", "undefined"},
			Fs:       files.NewFs(afero.NewOsFs(), scope, false),
			Rules:    []rules.Rule{{Allow: false, Path: "/private"}},
		},
	}
	return d, scope, outside
}

func TestUserCommandRunsOnlyTheDefinedAction(t *testing.T) {
	d, scope, _ := commandData(t)

	program, args, dir, err := userCommand(d, "/docs/", `ls "a b" ../docs/c /top`)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-la", "--", filepath.Join(scope, "docs", "a b"), filepath.Join(scope, "docs", "c"), filepath.Join(scope, "top")}
	if program != "/bin/ls" || !slices.Equal(args, want) {
		t.Errorf("got %s %q; want /bin/ls %q", program, args, want)
	}
	if dir != filepath.Join(scope, "docs") {
		t.Errorf("dir = %q", dir)
	}
}

// What the user types can only name files: options, absolute host paths and
// shell syntax all end up as the path of a file inside the scope.
func TestUserCommandArgumentsAreOnlyScopedPaths(t *testing.T) {
	d, scope, outside := commandData(t)

	for _, raw := range []string{
		"ls -R",
		"ls --recursive",
		"ls /etc/passwd",
		"ls ../../outside/secret",
		"ls ../../../../../../etc/passwd",
		"ls " + filepath.Join(outside, "secret"),
		"ls $(id)",
		"ls ;id",
		"ls `id`",
	} {
		_, args, _, err := userCommand(d, "/docs/", raw)
		if err != nil {
			continue // refused outright, which is safe too
		}
		if len(args) < 2 || args[0] != "-la" || args[1] != "--" {
			t.Errorf("%q: fixed arguments changed: %q", raw, args)
		}
		for _, arg := range args[2:] {
			if !strings.HasPrefix(arg, scope+string(filepath.Separator)) {
				t.Errorf("VULNERABLE: %q passed %q to the program", raw, arg)
			}
		}
	}
}

func TestUserCommandRefusals(t *testing.T) {
	d, _, _ := commandData(t)

	for _, tt := range []struct {
		name, cwd, raw string
		notAllowed     bool // the generic refusal, as opposed to a specific error
	}{
		{"not in the user's commands", "/", "secret", true},
		{"in the user's commands but not defined", "/", "undefined", true},
		{"a program instead of an action", "/", "/bin/sh -c id", true},
		{"action takes no arguments", "/", "pull origin", false},
		{"definition without an absolute program", "/", "relpath", false},
		{"symlink to a file outside the scope", "/docs/", "ls link", false},
		{"path through a symlink to a directory outside", "/", "ls dirlink/secret", false},
		{"path denied by a rule", "/", "ls private/x", false},
		{"working directory denied by a rule", "/private/", "ls", false},
		{"working directory behind an outside symlink", "/dirlink/", "ls", false},
		{"working directory does not exist", "/nope/", "ls", false},
		{"unterminated quote", "/", `ls "a`, false},
	} {
		program, _, _, err := userCommand(d, tt.cwd, tt.raw)
		if err == nil {
			t.Errorf("VULNERABLE: %s: %q would run %s", tt.name, tt.raw, program)
			continue
		}
		if errors.Is(err, errCommandNotAllowed) != tt.notAllowed {
			t.Errorf("%s: err = %v", tt.name, err)
		}
	}
}

// With followExternalSymlinks the operator has accepted symlinks that leave the
// scope, but names that climb out of it are still clamped to the scope.
func TestUserCommandFollowExternalSymlinks(t *testing.T) {
	d, scope, _ := commandData(t)
	d.user.Fs = files.NewFs(afero.NewOsFs(), scope, true)

	if _, _, _, err := userCommand(d, "/docs/", "ls link"); err != nil {
		t.Errorf("symlink refused although external symlinks are followed: %v", err)
	}
	_, args, _, err := userCommand(d, "/", "ls ../outside/secret")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := args[len(args)-1], filepath.Join(scope, "outside", "secret"); got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

// End to end over the websocket: an action runs in the user's directory and
// its output comes back, while a line that is not an action is refused.
func TestCommandsHandlerRunsAction(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX programs")
	}

	scope, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scope, "hello.txt"), []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{Execute: true}
	st := scopedUserStorage(t, scope, perm, key)
	if err := st.Settings.Save(&settings.Settings{Key: key, Actions: map[string]string{"ls": "/bin/ls -- $FILES"}}); err != nil {
		t.Fatal(err)
	}
	if err := st.Users.Update(&users.User{ID: 1, Commands: []string{"ls"}}, "Commands"); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(handle(commandsHandler, "", st, &settings.Server{EnableExec: true}))
	defer srv.Close()

	run := func(line string) string {
		t.Helper()
		header := http.Header{"X-Auth": {signToken(t, perm, key)}}
		conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/", header)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		defer conn.Close()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			t.Fatal(err)
		}
		var out []string
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return strings.Join(out, "\n")
			}
			out = append(out, string(msg))
		}
	}

	if got := run("ls"); got != "hello.txt" {
		t.Errorf("ls = %q; want hello.txt", got)
	}
	if got := run("ls hello.txt"); got != filepath.Join(scope, "hello.txt") {
		t.Errorf("ls hello.txt = %q", got)
	}
	for _, line := range []string{"/bin/sh -c id", "cat hello.txt", "sh"} {
		if got := run(line); got != string(cmdNotAllowed) {
			t.Errorf("%q = %q; want a refusal", line, got)
		}
	}
}
