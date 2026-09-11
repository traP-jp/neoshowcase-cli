package client

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSessionCookieTransport(t *testing.T) {
	t.Parallel()

	original, err := http.NewRequest(http.MethodGet, "https://ns.trap.jp", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	original.Header.Set("Cookie", "caller=value")

	transport := sessionCookieTransport{
		base: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if got, want := request.Header.Get("Cookie"), "session=value; other=value"; got != want {
				t.Errorf("Cookie header = %q, want %q", got, want)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
				Request:    request,
			}, nil
		}),
		sessionCookie: "session=value; other=value",
	}

	if _, err := transport.RoundTrip(original); err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if got, want := original.Header.Get("Cookie"), "caller=value"; got != want {
		t.Errorf("original Cookie header = %q, want %q", got, want)
	}
}
