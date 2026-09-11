package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/api/genconnect"
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
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

func newAPIClient(options model.Connection) apiClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: options.Insecure} //nolint:gosec // gated by an explicit dangerous flag
	client := &http.Client{Transport: authTransport{base: transport, header: options.AuthHeader, user: options.User}}
	return &connectAPIClient{APIServiceClient: genconnect.NewAPIServiceClient(client, strings.TrimRight(options.Endpoint, "/"))}
}

func rpcError(action string, err error) error {
	if connect.CodeOf(err) == connect.CodeNotFound {
		return cli.Fail(cli.ExitNotFound, "%s: target not found", action)
	}
	return fmt.Errorf("%s: %w", action, err)
}
