package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/api/genconnect"
)

type apiClient interface {
	GetApplications(context.Context, *connect.Request[api.GetApplicationsRequest]) (*connect.Response[api.GetApplicationsResponse], error)
	GetApplication(context.Context, *connect.Request[api.ApplicationIdRequest]) (*connect.Response[api.Application], error)
	GetOutput(context.Context, *connect.Request[api.GetOutputRequest]) (*connect.Response[api.ApplicationOutputs], error)
	GetOutputStream(context.Context, *connect.Request[api.GetOutputStreamRequest]) (messageStream[api.ApplicationOutput], error)
	StartApplication(context.Context, *connect.Request[api.ApplicationIdRequest]) (*connect.Response[emptypb.Empty], error)
	StopApplication(context.Context, *connect.Request[api.ApplicationIdRequest]) (*connect.Response[emptypb.Empty], error)
	GetAllBuilds(context.Context, *connect.Request[api.GetAllBuildsRequest]) (*connect.Response[api.GetBuildsResponse], error)
	GetBuilds(context.Context, *connect.Request[api.ApplicationIdRequest]) (*connect.Response[api.GetBuildsResponse], error)
	GetBuild(context.Context, *connect.Request[api.BuildIdRequest]) (*connect.Response[api.Build], error)
	RetryCommitBuild(context.Context, *connect.Request[api.RetryCommitBuildRequest]) (*connect.Response[emptypb.Empty], error)
	CancelBuild(context.Context, *connect.Request[api.BuildIdRequest]) (*connect.Response[emptypb.Empty], error)
	GetBuildLog(context.Context, *connect.Request[api.BuildIdRequest]) (*connect.Response[api.BuildLog], error)
	GetBuildLogStream(context.Context, *connect.Request[api.BuildIdRequest]) (messageStream[api.BuildLog], error)
}

type messageStream[T any] interface {
	Receive() bool
	Msg() *T
	Err() error
}

type connectAPIClient struct {
	genconnect.APIServiceClient
}

func (c *connectAPIClient) GetOutputStream(ctx context.Context, req *connect.Request[api.GetOutputStreamRequest]) (messageStream[api.ApplicationOutput], error) {
	return c.APIServiceClient.GetOutputStream(ctx, req)
}

func (c *connectAPIClient) GetBuildLogStream(ctx context.Context, req *connect.Request[api.BuildIdRequest]) (messageStream[api.BuildLog], error) {
	return c.APIServiceClient.GetBuildLogStream(ctx, req)
}

type authTransport struct {
	base   http.RoundTripper
	header string
	user   string
}

func (t authTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	clone.Header = request.Header.Clone()
	clone.Header.Set(t.header, t.user)
	return t.base.RoundTrip(clone)
}

func newAPIClient(cfg resolvedConfig) (apiClient, error) {
	parsed, err := url.Parse(cfg.endpoint)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fail(ExitUsage, "endpoint must be an http(s) URL without user info, a query, or a fragment")
	}
	if strings.TrimSpace(cfg.user) == "" {
		return nil, fail(ExitUsage, "NeoShowcase user is required (--user or NEOSHOWCASE_USER)")
	}
	if strings.ContainsAny(cfg.user, "\r\n") {
		return nil, fail(ExitUsage, "NeoShowcase user contains invalid characters")
	}
	header := textproto.CanonicalMIMEHeaderKey(cfg.authHeader)
	if !validHeaderName(header) {
		return nil, fail(ExitUsage, "invalid authentication header name")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: cfg.insecure} //nolint:gosec // gated by an explicit dangerous flag
	client := &http.Client{Transport: authTransport{base: transport, header: header, user: cfg.user}}
	return &connectAPIClient{APIServiceClient: genconnect.NewAPIServiceClient(client, strings.TrimRight(parsed.String(), "/"))}, nil
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

func rpcError(action string, err error) error {
	if connect.CodeOf(err) == connect.CodeNotFound {
		return fail(ExitNotFound, "%s: target not found", action)
	}
	return fmt.Errorf("%s: %w", action, err)
}
