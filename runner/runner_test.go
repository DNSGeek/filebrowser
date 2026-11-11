package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/DNSGeek/filebrowser/v3/settings"
	"github.com/DNSGeek/filebrowser/v3/users"
)

const hostileName = "x$(touch pwned)`touch pwned`;touch pwned"

func hookRunner(t *testing.T, s *settings.Settings) (r *Runner, user *users.User, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("requires a POSIX shell")
	}
	dir = t.TempDir()
	t.Chdir(dir)
	user = &users.User{Username: "u", Scope: "/", Fs: afero.NewBasePathFs(afero.NewOsFs(), dir)}
	return &Runner{Enabled: true, Settings: s}, user, dir
}

func assertNotPwned(t *testing.T, dir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, "pwned")); err == nil {
		t.Fatal("VULNERABLE: file name was executed")
	}
}

// A hook that calls a shell itself must not let the file name decide what the
// shell executes: $FILE inside the script is left for the shell to read from
// the environment, not pasted into the script text.
func TestHookScriptDoesNotInjectFileName(t *testing.T) {
	r, user, dir := hookRunner(t, &settings.Settings{
		Hooks: map[string][]string{"after_upload": {`sh -c 'printf %s "$FILE" > out; echo $FILE > /dev/null'`}},
	})

	if err := r.RunHook(func() error { return nil }, "upload", "/"+hostileName, "", user); err != nil {
		t.Fatal(err)
	}

	assertNotPwned(t, dir)
	got, _ := os.ReadFile(filepath.Join(dir, "out"))
	if want := filepath.Join(dir, hostileName); string(got) != want {
		t.Errorf("$FILE = %q; want %q", got, want)
	}
}

// An argument that is exactly a hook variable is replaced by its value and
// reaches the program as one argument, whatever characters it holds.
func TestHookExpandsWholeArguments(t *testing.T) {
	r, user, dir := hookRunner(t, &settings.Settings{
		Hooks: map[string][]string{"after_upload": {`sh -c 'printf "%s|%s|%s" "$1" "$2" "$3" > out' hook $FILE ${TRIGGER} $HOME`}},
	})

	name := hostileName + " with spaces"
	if err := r.RunHook(func() error { return nil }, "upload", "/"+name, "", user); err != nil {
		t.Fatal(err)
	}

	assertNotPwned(t, dir)
	got, _ := os.ReadFile(filepath.Join(dir, "out"))
	// $HOME is not a hook variable and stays as written.
	if want := filepath.Join(dir, name) + "|after_upload|$HOME"; string(got) != want {
		t.Errorf("arguments = %q; want %q", got, want)
	}
}

// A hook holding shell syntax is run as a plain program with plain arguments.
func TestHookIsNotRunThroughAShell(t *testing.T) {
	r, user, dir := hookRunner(t, &settings.Settings{
		Hooks: map[string][]string{"after_upload": {`true ; touch pwned`}},
	})

	if err := r.RunHook(func() error { return nil }, "upload", "/f", "", user); err != nil {
		t.Fatal(err)
	}
	assertNotPwned(t, dir)
}

// Hooks left in the old, web-editable "commands" setting are never run.
func TestLegacyCommandsAreNotRun(t *testing.T) {
	r, user, dir := hookRunner(t, &settings.Settings{
		Commands: map[string][]string{"after_upload": {`touch pwned`}, "before_upload": {`touch pwned`}},
	})

	if err := r.RunHook(func() error { return nil }, "upload", "/f", "", user); err != nil {
		t.Fatal(err)
	}
	assertNotPwned(t, dir)
}

func TestDisabledRunnerRunsNoHooks(t *testing.T) {
	r, user, dir := hookRunner(t, &settings.Settings{
		Hooks: map[string][]string{"after_upload": {`touch pwned`}},
	})
	r.Enabled = false

	if err := r.RunHook(func() error { return nil }, "upload", "/f", "", user); err != nil {
		t.Fatal(err)
	}
	assertNotPwned(t, dir)
}

// A value that would be read as an option is refused rather than passed on.
func TestHookRefusesOptionLikeValue(t *testing.T) {
	r, user, dir := hookRunner(t, &settings.Settings{
		Hooks: map[string][]string{"before_upload": {`touch $USERNAME`}},
	})
	user.Username = "--reference=pwned"

	ran := false
	err := r.RunHook(func() error { ran = true; return nil }, "upload", "/f", "", user)
	if err == nil || !strings.Contains(err.Error(), "USERNAME") {
		t.Fatalf("err = %v; want a refusal naming USERNAME", err)
	}
	if ran {
		t.Error("the action ran although its before hook failed")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("hook ran: %v", entries)
	}
}

func TestExpandHookArg(t *testing.T) {
	values := map[string]string{"FILE": "/a b", "USERNAME": "-x"}
	for arg, want := range map[string]string{
		"$FILE":       "/a b",
		"${FILE}":     "/a b",
		"x$FILE":      "x$FILE",
		"$FILE.bak":   "$FILE.bak",
		"--out=$FILE": "--out=$FILE",
		"$PATH":       "$PATH",
		"${FILE":      "${FILE",
		"$":           "$",
		"plain":       "plain",
	} {
		got, err := expandHookArg(arg, values)
		if err != nil || got != want {
			t.Errorf("expandHookArg(%q) = %q, %v; want %q", arg, got, err, want)
		}
	}
	if _, err := expandHookArg("$USERNAME", values); err == nil {
		t.Error("expandHookArg($USERNAME) with value -x: want an error")
	}
}
