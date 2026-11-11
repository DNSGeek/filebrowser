package runner

import (
	"errors"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func abs(p string) string {
	if runtime.GOOS == "windows" {
		return `C:` + filepath.FromSlash(p)
	}
	return p
}

func TestValidateAction(t *testing.T) {
	ls := abs("/bin/ls")
	for _, tt := range []struct {
		name, definition string
		ok               bool
	}{
		{"ls", ls + " -la -- $FILES", true},
		{"git-pull", abs("/usr/bin/git") + " pull", true},
		{"ls", "ls -la", false},               // PATH lookup
		{"ls", "./ls", false},                 // relative to the user's directory
		{"ls", ls + " --color=$FILES", false}, // paths inside another argument
		{"ls", ls + " $FILES $FILES", false},  // more than one placeholder
		{"ls", "", false},                     // nothing to run
		{"", ls, false},                       // no name
		{"-rf", ls, false},                    // option-like name
		{"a b", ls, false},                    // not a single word
		{"../x", ls, false},                   // not a plain name
	} {
		if err := ValidateAction(tt.name, tt.definition); (err == nil) != tt.ok {
			t.Errorf("ValidateAction(%q, %q) = %v; want ok=%v", tt.name, tt.definition, err, tt.ok)
		}
	}
}

func TestBuildAction(t *testing.T) {
	ls := abs("/bin/ls")
	a, b := abs("/srv/u/a b"), abs("/srv/u/$(touch pwned);x")

	program, args, err := BuildAction("ls", ls+" -la -- $FILES", []string{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"-la", "--", a, b}; program != ls || !slices.Equal(args, want) {
		t.Errorf("got %q %q; want %q %q", program, args, ls, want)
	}

	// No paths: the placeholder disappears.
	if _, args, err = BuildAction("ls", ls+" -la -- $FILES", nil); err != nil || !slices.Equal(args, []string{"-la", "--"}) {
		t.Errorf("no paths: got %q, %v", args, err)
	}

	// An action without a placeholder accepts nothing from the user.
	if _, _, err = BuildAction("pull", abs("/usr/bin/git")+" pull", []string{a}); !errors.Is(err, ErrActionTakesNoFiles) {
		t.Errorf("err = %v; want ErrActionTakesNoFiles", err)
	}

	// Only absolute paths are accepted, so nothing can be read as an option.
	for _, p := range []string{"-rf", "--exec=sh", "relative/path", ""} {
		if _, _, err = BuildAction("ls", ls+" $FILES", []string{p}); err == nil {
			t.Errorf("path %q was accepted", p)
		}
	}

	// A definition that is no longer valid is not run.
	if _, _, err = BuildAction("ls", "ls $FILES", nil); err == nil {
		t.Error("relative program was accepted")
	}
}
