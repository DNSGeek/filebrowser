package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/DNSGeek/filebrowser/v3/settings"
	"github.com/DNSGeek/filebrowser/v3/users"
)

// Runner runs the hooks configured for file events.
type Runner struct {
	Enabled bool
	*settings.Settings
}

// RunHook runs the hooks for the before and after event.
func (r *Runner) RunHook(fn func() error, evt, path, dst string, user *users.User) error {
	path = user.FullPath(path)
	dst = user.FullPath(dst)

	if r.Enabled {
		if val, ok := r.Hooks["before_"+evt]; ok {
			for _, command := range val {
				err := r.exec(command, "before_"+evt, path, dst, user)
				if err != nil {
					return err
				}
			}
		}
	}

	err := fn()
	if err != nil {
		return err
	}

	if r.Enabled {
		if val, ok := r.Hooks["after_"+evt]; ok {
			for _, command := range val {
				err := r.exec(command, "after_"+evt, path, dst, user)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// hookVars are the variables a hook argument may refer to.
var hookVars = []string{"FILE", "SCOPE", "TRIGGER", "USERNAME", "DESTINATION"}

// expandHookArg replaces an argument that is exactly "$NAME" or "${NAME}", for
// one of the hook variables, with that variable's value.
//
// Only whole arguments are replaced. Values such as file names are chosen by
// users, so they are never pasted into the middle of an argument: in
// `sh -c 'echo $FILE'` that would let a file name be parsed as shell syntax.
// Such a script still works, and safely, because the same variables are set in
// the command's environment.
//
// A value starting with "-" is refused, since the command would read it as one
// of its options instead of as data.
func expandHookArg(arg string, values map[string]string) (string, error) {
	name, ok := strings.CutPrefix(arg, "$")
	if !ok {
		return arg, nil
	}
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		name = name[1 : len(name)-1]
	}

	value, ok := values[name]
	if !ok {
		return arg, nil
	}
	if strings.HasPrefix(value, "-") {
		return "", fmt.Errorf("hook: refusing to pass %s=%q as an argument: it would be read as an option", name, value)
	}
	return value, nil
}

// exec runs one hook. The hook is split into a program and its arguments and
// started directly: it never goes through a shell, whatever the "shell"
// setting says, so nothing in it or in the values it receives is interpreted
// as shell syntax.
func (r *Runner) exec(raw, evt, path, dst string, user *users.User) error {
	blocking := true

	if strings.HasSuffix(raw, "&") {
		blocking = false
		raw = strings.TrimSpace(strings.TrimSuffix(raw, "&"))
	}

	name, args, err := SplitCommandAndArgs(raw)
	if err != nil {
		return err
	}

	values := map[string]string{
		"FILE":        path,
		"SCOPE":       user.Scope,
		"TRIGGER":     evt,
		"USERNAME":    user.Username,
		"DESTINATION": dst,
	}

	for i, arg := range args {
		args[i], err = expandHookArg(arg, values)
		if err != nil {
			return err
		}
	}

	command := append([]string{name}, args...)

	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	for _, k := range hookVars {
		cmd.Env = append(cmd.Env, k+"="+values[k])
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if !blocking {
		log.Printf("[INFO] Nonblocking Command: \"%s\"", strings.Join(command, " "))
		defer func() {
			go func() {
				err := cmd.Wait()
				if err != nil {
					log.Printf("[INFO] Nonblocking Command \"%s\" failed: %s", strings.Join(command, " "), err)
				}
			}()
		}()
		return cmd.Start()
	}

	log.Printf("[INFO] Blocking Command: \"%s\"", strings.Join(command, " "))
	return cmd.Run()
}
