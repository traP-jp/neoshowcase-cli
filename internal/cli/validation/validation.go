package validation

import (
	"fmt"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

const defaultAuthHeader = "X-Showcase-User"

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

func Connection(endpoint, user, authHeader string, insecure bool) (model.Connection, error) {
	endpoint = strings.TrimSpace(endpoint)
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return model.Connection{}, fmt.Errorf("endpoint must be an http(s) URL without user info, a query, or a fragment")
	}
	user = strings.TrimSpace(user)
	if user == "" {
		return model.Connection{}, fmt.Errorf("NeoShowcase user is required (NEOSHOWCASE_USER)")
	}
	if strings.ContainsAny(user, "\r\n") {
		return model.Connection{}, fmt.Errorf("NeoShowcase user contains invalid characters")
	}
	authHeader = textproto.CanonicalMIMEHeaderKey(firstNonEmpty(authHeader, defaultAuthHeader))
	if !validHeaderName(authHeader) {
		return model.Connection{}, fmt.Errorf("invalid authentication header name")
	}
	return model.Connection{Endpoint: endpoint, User: user, AuthHeader: authHeader, Insecure: insecure}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func validHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			continue
		}
		switch c {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		default:
			return false
		}
	}
	return true
}
