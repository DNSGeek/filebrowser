package runner

import (
	"errors"
	"strings"

	"github.com/filebrowser/filebrowser/v2/settings"
)

// shellMetachars are the characters a shell (sh, bash, cmd.exe, PowerShell)
// treats as command separators, redirections, substitutions, subexpressions or
// variable expansions. A user-issued command containing any of them could run
// more than the single allowed command when it is handed to a shell.
const shellMetachars = ";&|<>()$`^%\n\r"

// ErrShellMetachars is returned when a user-issued command contains shell
// metacharacters while the instance is configured to run commands through a
// shell.
var ErrShellMetachars = errors.New("command contains characters that are not allowed: " +
	"; & | < > ( ) $ ` ^ % and line breaks")

// ParseCommand parses the command taking in account if the current
// instance uses a shell to run the commands or just calls the binary
// directly.
func ParseCommand(s *settings.Settings, raw string) (command []string, name string, err error) {
	name, args, err := SplitCommandAndArgs(raw)
	if err != nil {
		return
	}

	if usesShell(s) {
		command = append(command, s.Shell...)
		command = append(command, raw)
	} else {
		command = append(command, name)
		command = append(command, args...)
	}

	return command, name, nil
}

// ParseUserCommand parses a command issued by a user through the command
// runner and checks it against the user's allowed commands.
//
// Only the command name is checked against allowed, so when commands run
// through a shell the rest of the line must not be able to start another
// command (e.g. "ls; rm -rf ~" or "ls $(rm -rf ~)"). Such lines are rejected
// with ErrShellMetachars. When no shell is used, the executable in the returned
// command is taken from allowed rather than from raw.
func ParseUserCommand(s *settings.Settings, raw string, allowed []string) (command []string, allowedName bool, err error) {
	if usesShell(s) && strings.ContainsAny(raw, shellMetachars) {
		return nil, false, ErrShellMetachars
	}

	command, name, err := ParseCommand(s, raw)
	if err != nil {
		return nil, false, err
	}

	for _, a := range allowed {
		if a == name {
			if !usesShell(s) {
				command[0] = a
			}
			return command, true, nil
		}
	}

	return nil, false, nil
}

func usesShell(s *settings.Settings) bool {
	return len(s.Shell) > 0 && s.Shell[0] != ""
}
