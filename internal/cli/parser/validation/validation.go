package validation

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func Mutable(allowed bool) error {
	if !allowed {
		return fmt.Errorf("this command changes NeoShowcase state; pass --allow-mutable-operation for this invocation")
	}
	return nil
}

func Timeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("--timeout must be greater than zero")
	}
	return nil
}

func Connection(endpoint, sessionCookie string) (model.Connection, error) {
	endpoint = strings.TrimSpace(endpoint)
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return model.Connection{}, fmt.Errorf("endpoint must be an http(s) URL without user info, a query, or a fragment")
	}
	sessionCookie = strings.TrimSpace(sessionCookie)
	if sessionCookie == "" {
		return model.Connection{}, fmt.Errorf("NeoShowcase session cookie is required (NEOSHOWCASE_SESSION_COOKIE)")
	}
	if strings.ContainsAny(sessionCookie, "\r\n") {
		return model.Connection{}, fmt.Errorf("NeoShowcase session cookie contains invalid characters")
	}
	return model.Connection{Endpoint: endpoint, SessionCookie: sessionCookie}, nil
}
