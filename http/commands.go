package fbhttp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/DNSGeek/filebrowser/v3/files"
	"github.com/DNSGeek/filebrowser/v3/runner"
)

const (
	WSWriteDeadline = 10 * time.Second

	// commandTimeout bounds how long a command run from the web UI may live.
	commandTimeout = 10 * time.Minute

	// maxRunningCommands bounds how many commands run from the web UI at once.
	maxRunningCommands = 8
)

var runningCommands = make(chan struct{}, maxRunningCommands)

// commandEnvKeys are the only variables a command run from the web UI inherits
// from the server. The rest of the environment may hold secrets (credentials,
// tokens, FB_* settings) that the people allowed to run commands must not read.
var commandEnvKeys = []string{
	"PATH", "HOME", "USER", "LOGNAME", "SHELL", "LANG", "LANGUAGE", "TZ", "TMPDIR", "TERM",
	"SYSTEMROOT", "COMSPEC", "PATHEXT", "WINDIR", "TEMP", "TMP",
}

func commandEnv() []string {
	var env []string
	for _, k := range commandEnvKeys {
		if v, ok := os.LookupEnv(k); ok {
			env = append(env, k+"="+v)
		}
	}
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "LC_") {
			env = append(env, kv)
		}
	}
	return env
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	cmdNotAllowed = []byte("Command not allowed.")
)

var errCommandNotAllowed = errors.New("command not allowed")

// userCommand turns the line a user typed into the program to run, its
// arguments and its working directory.
//
// The first word names an action. It must be defined by the operator (see
// "filebrowser actions") and listed in the user's commands. Every other word
// is taken as a path, relative to cwd, and is passed on only as the real path
// of a file the user is allowed to reach: inside their scope, not behind a
// symbolic link leading out of it, and not denied by a rule. Nothing else the
// user typed reaches the program, so they can neither add options nor name
// files outside their scope.
func userCommand(d *data, cwd, raw string) (program string, args []string, dir string, err error) {
	name, words, err := runner.SplitCommandAndArgs(raw)
	if err != nil {
		return "", nil, "", err
	}

	definition, defined := d.settings.Actions[name]
	if !defined || !slices.Contains(d.user.Commands, name) {
		return "", nil, "", errCommandNotAllowed
	}

	dir, err = commandPath(d, "/", cwd)
	if err != nil {
		return "", nil, "", err
	}
	if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
		return "", nil, "", fmt.Errorf("%s: not a directory", cwd)
	}

	paths := make([]string, 0, len(words))
	for _, word := range words {
		p, pathErr := commandPath(d, cwd, word)
		if pathErr != nil {
			return "", nil, "", pathErr
		}
		paths = append(paths, p)
	}

	program, args, err = runner.BuildAction(name, definition, paths)
	if err != nil {
		return "", nil, "", err
	}
	return program, args, dir, nil
}

// commandPath resolves name, relative to the directory cwd of the user's
// scope, to its real path on disk, or fails if the user may not reach it.
func commandPath(d *data, cwd, name string) (string, error) {
	p := slashClean(name)
	if !strings.HasPrefix(filepath.ToSlash(name), "/") {
		p = slashClean(path.Join(cwd, filepath.ToSlash(name)))
	}

	if !d.Check(p) {
		return "", fmt.Errorf("%s: not allowed", name)
	}

	onDisk, err := files.GuardedRealPath(d.user.Fs, p)
	if err != nil {
		return "", fmt.Errorf("%s: not allowed", name)
	}
	return onDisk, nil
}

func wsErr(ws *websocket.Conn, r *http.Request, status int, err error) {
	txt := http.StatusText(status)
	if err != nil || status >= 400 {
		log.Printf("%s: %v %s %v", r.URL.Path, status, r.RemoteAddr, err)
	}
	if err := ws.WriteControl(websocket.CloseInternalServerErr, []byte(txt), time.Now().Add(WSWriteDeadline)); err != nil {
		log.Print(err)
	}
}

var commandsHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer conn.Close()

	var raw string

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
			return 0, nil
		}

		raw = strings.TrimSpace(string(msg))
		if raw != "" {
			break
		}
	}

	// Fail fast
	if !d.server.EnableExec || !d.user.Perm.Execute {
		if err := conn.WriteMessage(websocket.TextMessage, cmdNotAllowed); err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
		}

		return 0, nil
	}

	program, args, dir, err := userCommand(d, requestPath(r), raw)
	if err != nil {
		msg := cmdNotAllowed
		if !errors.Is(err, errCommandNotAllowed) {
			msg = []byte(err.Error())
		}
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
		}
		return 0, nil
	}

	select {
	case runningCommands <- struct{}{}:
		defer func() { <-runningCommands }()
	default:
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Too many commands are running.")); err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
		}
		return 0, nil
	}

	// The command ends with the timeout or as soon as the client goes away,
	// instead of running on, unobserved, on the server.
	ctx, cancel := context.WithTimeout(r.Context(), commandTimeout)
	defer cancel()

	go func() {
		// The client sends nothing after the command; this read only returns
		// when the connection closes.
		_, _, _ = conn.ReadMessage()
		cancel()
	}()

	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Env = commandEnv()
	// Killing a shell leaves its children holding the output pipes open.
	cmd.WaitDelay = 5 * time.Second
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return 0, nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return 0, nil
	}

	if err := cmd.Start(); err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return 0, nil
	}

	s := bufio.NewScanner(io.MultiReader(stdout, stderr))
	for s.Scan() {
		if err := conn.WriteMessage(websocket.TextMessage, s.Bytes()); err != nil {
			log.Print(err)
		}
	}

	if err := cmd.Wait(); err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
	}

	return 0, nil
})
