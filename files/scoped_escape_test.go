package files

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

// TestNewFsRejectsSiblingPrefixEscape guards against names that climb out of
// the scope into a sibling directory whose name starts with the scope's name
// (e.g. "/../user2" from "/srv/user"). afero's BasePathFs accepts those because
// it compares paths with a plain string prefix, so the check has to happen in
// ScopedFs — in both symlink modes.
func TestNewFsRejectsSiblingPrefixEscape(t *testing.T) {
	for _, followExternal := range []bool{false, true} {
		t.Run(map[bool]string{false: "scoped", true: "followExternal"}[followExternal], func(t *testing.T) {
			root := t.TempDir()
			scope := filepath.Join(root, "user")
			sibling := filepath.Join(root, "user2")
			for _, dir := range []string{scope, sibling} {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			secret := filepath.Join(sibling, "secret")
			if err := os.WriteFile(secret, []byte("secret"), 0o600); err != nil {
				t.Fatal(err)
			}

			fsys := NewFs(afero.NewOsFs(), scope, followExternal)

			if _, err := afero.ReadFile(fsys, "/../user2/secret"); !errors.Is(err, os.ErrPermission) {
				t.Errorf("read: got %v, want permission error", err)
			}
			if err := afero.WriteFile(fsys, "/../user2/pwned", []byte("x"), 0o600); !errors.Is(err, os.ErrPermission) {
				t.Errorf("write: got %v, want permission error", err)
			}
			if _, err := os.Stat(filepath.Join(sibling, "pwned")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("write escaped the scope: stat err = %v", err)
			}
			if err := fsys.MkdirAll("/../user2/dir", 0o755); !errors.Is(err, os.ErrPermission) {
				t.Errorf("mkdir: got %v, want permission error", err)
			}
			if err := fsys.Rename("/../user2/secret", "/stolen"); !errors.Is(err, os.ErrPermission) {
				t.Errorf("rename: got %v, want permission error", err)
			}
			if err := fsys.RemoveAll("/../user2"); !errors.Is(err, os.ErrPermission) {
				t.Errorf("remove: got %v, want permission error", err)
			}
			if _, err := os.Stat(secret); err != nil {
				t.Errorf("sibling file was touched: %v", err)
			}
			if _, err := fsys.(*ScopedFs).RealPath("/../user2/secret"); !errors.Is(err, os.ErrPermission) {
				t.Errorf("RealPath: got %v, want permission error", err)
			}

			// Ordinary names inside the scope keep working.
			if err := afero.WriteFile(fsys, "/a/../ok.txt", []byte("ok"), 0o600); err != nil {
				t.Errorf("write inside scope: %v", err)
			}
			if _, err := os.Stat(filepath.Join(scope, "ok.txt")); err != nil {
				t.Errorf("file inside scope missing: %v", err)
			}
		})
	}
}

// TestNewFsNestedShareRejectsEscape covers a share rebased onto a subfolder of
// the user's scope: a name must not climb from the share into a sibling folder
// of the same user.
func TestNewFsNestedShareRejectsEscape(t *testing.T) {
	for _, followExternal := range []bool{false, true} {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, "docs-private"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "docs-private", "secret"), []byte("s"), 0o600); err != nil {
			t.Fatal(err)
		}

		userFs := NewFs(afero.NewOsFs(), root, followExternal)
		shareFs := NewFs(userFs, "/docs", followExternal)

		if _, err := afero.ReadFile(shareFs, "/../docs-private/secret"); !errors.Is(err, os.ErrPermission) {
			t.Errorf("followExternal=%v: got %v, want permission error", followExternal, err)
		}
	}
}
