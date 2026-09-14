package runner

import (
	"errors"
	"slices"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
)

func TestParseUserCommand(t *testing.T) {
	shell := &settings.Settings{Shell: []string{"sh", "-c"}}
	noShell := &settings.Settings{}
	allowed := []string{"ls", "git"}

	tests := []struct {
		name        string
		settings    *settings.Settings
		raw         string
		wantCommand []string
		wantAllowed bool
		wantErr     error
	}{
		{"shell allowed", shell, "ls -la 'my dir'", []string{"sh", "-c", "ls -la 'my dir'"}, true, nil},
		{"shell glob allowed", shell, "ls *.txt", []string{"sh", "-c", "ls *.txt"}, true, nil},
		{"shell not allowed", shell, "rm -rf x", nil, false, nil},
		{"shell semicolon", shell, "ls; rm -rf ~", nil, false, ErrShellMetachars},
		{"shell and", shell, "ls && rm -rf ~", nil, false, ErrShellMetachars},
		{"shell pipe", shell, "ls | sh", nil, false, ErrShellMetachars},
		{"shell background", shell, "ls & rm -rf ~", nil, false, ErrShellMetachars},
		{"shell substitution", shell, "ls $(rm -rf ~)", nil, false, ErrShellMetachars},
		{"shell backticks", shell, "ls `rm -rf ~`", nil, false, ErrShellMetachars},
		{"shell variable", shell, "ls $HOME", nil, false, ErrShellMetachars},
		{"shell redirect", shell, "ls > /etc/passwd", nil, false, ErrShellMetachars},
		{"shell subexpression", shell, "ls (rm x)", nil, false, ErrShellMetachars},
		{"shell newline", shell, "ls\nrm -rf ~", nil, false, ErrShellMetachars},
		{"shell cmd escape", shell, "ls ^& del x", nil, false, ErrShellMetachars},
		{"no shell allowed", noShell, "git log -1", []string{"git", "log", "-1"}, true, nil},
		{"no shell metachars are plain args", noShell, "ls ;", []string{"ls", ";"}, true, nil},
		{"no shell not allowed", noShell, "rm x", nil, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, ok, err := ParseUserCommand(tt.settings, tt.raw, allowed)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if ok != tt.wantAllowed {
				t.Fatalf("allowed = %v, want %v", ok, tt.wantAllowed)
			}
			if !slices.Equal(command, tt.wantCommand) {
				t.Fatalf("command = %q, want %q", command, tt.wantCommand)
			}
		})
	}
}
