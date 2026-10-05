package fbhttp

import (
	"strings"
	"testing"
)

func TestCommandEnvOmitsServerSecrets(t *testing.T) {
	t.Setenv("FB_PASSWORD", "hunter2")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "s3cr3t")
	t.Setenv("PATH", "/usr/bin")
	t.Setenv("LC_ALL", "C")

	env := strings.Join(commandEnv(), "\n")
	for _, secret := range []string{"FB_PASSWORD", "AWS_SECRET_ACCESS_KEY", "hunter2"} {
		if strings.Contains(env, secret) {
			t.Errorf("command environment leaks %s", secret)
		}
	}
	for _, want := range []string{"PATH=/usr/bin", "LC_ALL=C"} {
		if !strings.Contains(env, want) {
			t.Errorf("command environment lacks %s", want)
		}
	}
}
