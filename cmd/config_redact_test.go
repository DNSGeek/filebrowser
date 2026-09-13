package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/auth"
)

func TestRedactAutherMasksReCaptchaSecret(t *testing.T) {
	original := &auth.JSONAuth{ReCaptcha: &auth.ReCaptcha{
		Host:   "https://www.google.com",
		Key:    "site-key",
		Secret: "super-secret",
	}}

	b, err := json.Marshal(redactAuther(original))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "super-secret") {
		t.Fatalf("secret leaked in printed config: %s", b)
	}
	if !strings.Contains(string(b), "site-key") {
		t.Fatalf("non-secret fields should still be printed: %s", b)
	}
	if original.ReCaptcha.Secret != "super-secret" {
		t.Fatal("redactAuther modified the stored auther")
	}
}

func TestRedactAutherLeavesOtherAuthersAlone(t *testing.T) {
	proxy := &auth.ProxyAuth{Header: "X-User"}
	if got := redactAuther(proxy); got != proxy {
		t.Fatalf("got %#v, want the original proxy auther", got)
	}
	if got := redactAuther(&auth.JSONAuth{}); got.(*auth.JSONAuth).ReCaptcha != nil {
		t.Fatalf("unexpected recaptcha on empty JSON auther: %#v", got)
	}
}
