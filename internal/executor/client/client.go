package client

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
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

type Client interface {
	GetApplications(context.Context, *connect.Request[api.GetApplicationsRequest]) (*connect.Response[api.GetApplicationsResponse], error)
	GetApplication(context.Context, *connect.Request[api.ApplicationIdRequest]) (*connect.Response[api.Application], error)
	GetOutput(context.Context, *connect.Request[api.GetOutputRequest]) (*connect.Response[api.ApplicationOutputs], error)
	GetOutputStream(context.Context, *connect.Request[api.GetOutputStreamRequest]) (MessageStream[api.ApplicationOutput], error)
	StartApplication(context.Context, *connect.Request[api.ApplicationIdRequest]) (*connect.Response[emptypb.Empty], error)
	StopApplication(context.Context, *connect.Request[api.ApplicationIdRequest]) (*connect.Response[emptypb.Empty], error)
	GetAllBuilds(context.Context, *connect.Request[api.GetAllBuildsRequest]) (*connect.Response[api.GetBuildsResponse], error)
	GetBuilds(context.Context, *connect.Request[api.ApplicationIdRequest]) (*connect.Response[api.GetBuildsResponse], error)
	GetBuild(context.Context, *connect.Request[api.BuildIdRequest]) (*connect.Response[api.Build], error)
	RetryCommitBuild(context.Context, *connect.Request[api.RetryCommitBuildRequest]) (*connect.Response[emptypb.Empty], error)
	CancelBuild(context.Context, *connect.Request[api.BuildIdRequest]) (*connect.Response[emptypb.Empty], error)
	GetBuildLog(context.Context, *connect.Request[api.BuildIdRequest]) (*connect.Response[api.BuildLog], error)
	GetBuildLogStream(context.Context, *connect.Request[api.BuildIdRequest]) (MessageStream[api.BuildLog], error)
}

type MessageStream[T any] interface {
	Receive() bool
	Msg() *T
	Err() error
}

type connectClient struct {
	genconnect.APIServiceClient
}

func (c *connectClient) GetOutputStream(ctx context.Context, request *connect.Request[api.GetOutputStreamRequest]) (MessageStream[api.ApplicationOutput], error) {
	return c.APIServiceClient.GetOutputStream(ctx, request)
}

func (c *connectClient) GetBuildLogStream(ctx context.Context, request *connect.Request[api.BuildIdRequest]) (MessageStream[api.BuildLog], error) {
	return c.APIServiceClient.GetBuildLogStream(ctx, request)
}

type sessionCookieTransport struct {
	base          http.RoundTripper
	sessionCookie string
}

func (t sessionCookieTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	clone.Header = request.Header.Clone()
	clone.Header.Set("Cookie", t.sessionCookie)
	return t.base.RoundTrip(clone)
}

func New(options model.Connection) Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	httpClient := &http.Client{Transport: sessionCookieTransport{base: transport, sessionCookie: options.SessionCookie}}
	return &connectClient{APIServiceClient: genconnect.NewAPIServiceClient(httpClient, strings.TrimRight(options.Endpoint, "/"))}
}

func ResolveApplication(ctx context.Context, apiClient Client, identifier string) (*api.Application, error) {
	response, err := apiClient.GetApplications(ctx, connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
	if err != nil {
		return nil, RPCError("resolve application", err)
	}
	var matches []*api.Application
	for _, application := range response.Msg.GetApplications() {
		if application.GetId() == identifier {
			return application, nil
		}
		if application.GetName() == identifier {
			matches = append(matches, application)
		}
	}
	if len(matches) == 0 {
		return nil, model.NewError(model.ErrorNotFound, "application %q not found", identifier)
	}
	if len(matches) > 1 {
		return nil, model.NewError(model.ErrorNotFound, "application name %q is ambiguous (%d exact matches)", identifier, len(matches))
	}
	return matches[0], nil
}

func ContextError(ctx context.Context, err error) error {
	if ctx.Err() == context.DeadlineExceeded {
		return model.NewError(model.ErrorTimeout, "command timed out")
	}
	if ctx.Err() == context.Canceled {
		return model.NewError(model.ErrorInterrupt, "command interrupted")
	}
	return err
}

func RPCError(action string, err error) error {
	if connect.CodeOf(err) == connect.CodeNotFound {
		return model.NewError(model.ErrorNotFound, "%s: target not found", action)
	}
	return fmt.Errorf("%s: %w", action, err)
}
