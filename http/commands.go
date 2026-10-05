package fbhttp

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"

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

	program, args, allowed, err := runner.ParseUserCommand(d.settings, raw, d.user.Commands)
	if err != nil {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
		}
		return 0, nil
	}

	if !allowed {
		if err := conn.WriteMessage(websocket.TextMessage, cmdNotAllowed); err != nil {
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
	cmd.Dir = d.user.FullPath(requestPath(r))

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
