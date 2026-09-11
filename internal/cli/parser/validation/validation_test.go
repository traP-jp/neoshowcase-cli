package validation

import (
	"strings"
	"testing"
)

func TestConnection(t *testing.T) {
	t.Parallel()

	connection, err := Connection(" https://ns.trap.jp/ ", " session=value; other=value ")
	if err != nil {
		t.Fatalf("Connection() error = %v", err)
	}
	if got, want := connection.Endpoint, "https://ns.trap.jp/"; got != want {
		t.Errorf("endpoint = %q, want %q", got, want)
	}
	if got, want := connection.SessionCookie, "session=value; other=value"; got != want {
		t.Errorf("session cookie = %q, want %q", got, want)
	}
}

func TestConnectionRequiresSessionCookie(t *testing.T) {
	t.Parallel()

	_, err := Connection("https://ns.trap.jp", "")
	if err == nil || !strings.Contains(err.Error(), "NEOSHOWCASE_SESSION_COOKIE") {
		t.Fatalf("Connection() error = %v, want missing session cookie error", err)
	}
}

func TestConnectionRejectsSessionCookieWithNewline(t *testing.T) {
	t.Parallel()

	_, err := Connection("https://ns.trap.jp", "session=value\r\nX-Showcase-User: attacker")
	if err == nil {
		t.Fatal("Connection() error = nil, want invalid session cookie error")
	}
}
