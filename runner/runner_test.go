package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// A hook run through a shell must not let the file name decide what the shell
// executes: $FILE and friends reach the script as environment variables, not
// as text pasted into the command line.
func TestShellHookDoesNotInjectFileName(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires a POSIX shell")
	}

	dir := t.TempDir()
	pwned := filepath.Join(dir, "pwned")
	out := filepath.Join(dir, "out")

	r := &Runner{
		Enabled: true,
		Settings: &settings.Settings{
			Shell:    []string{"sh", "-c"},
			Commands: map[string][]string{"after_upload": {`printf '%s' "$FILE" > ` + out}},
		},
	}
	user := &users.User{Username: "u", Scope: "/", Fs: afero.NewBasePathFs(afero.NewOsFs(), dir)}

	name := "x$(touch " + pwned + ")`touch " + pwned + "`;touch " + pwned
	if err := r.RunHook(func() error { return nil }, "upload", "/"+name, "", user); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(pwned); err == nil {
		t.Fatal("VULNERABLE: file name was executed by the shell")
	}
	got, _ := os.ReadFile(out)
	if want := filepath.Join(dir, name); string(got) != want {
		t.Errorf("$FILE = %q; want %q", got, want)
	}
}
