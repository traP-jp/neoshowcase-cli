package cli

import (
	"context"
	"time"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
)

type ConnectionOptions struct {
	Endpoint   string
	User       string
	AuthHeader string
	Insecure   bool
}

type MutationResult struct {
	Operation       string
	ApplicationID   string
	ApplicationName string
	BuildID         string
	SourceBuildID   string
	Commit          string
	Status          string
	State           string
}

type Output interface {
	Applications([]*api.Application) error
	Application(*api.Application) error
	ApplicationLogs(string, []*api.ApplicationOutput, bool) error
	Builds([]*api.Build, map[string]string) error
	BuildDetails(*api.Build) error
	BuildResult(*api.Build, string, bool) error
	BuildLog(string, []byte, bool) error
	Mutation(MutationResult) error
}

type Executor interface {
	AppList(context.Context, ConnectionOptions) error
	AppGet(context.Context, ConnectionOptions, string) error
	AppLogs(context.Context, ConnectionOptions, string, bool, int32, time.Time) error
	AppStart(context.Context, ConnectionOptions, string) error
	AppStop(context.Context, ConnectionOptions, string) error
	AppRestart(context.Context, ConnectionOptions, string) error
	AppRebuild(context.Context, ConnectionOptions, string, string, bool, bool) error
	BuildList(context.Context, ConnectionOptions, string, int32, int32) error
	BuildGet(context.Context, ConnectionOptions, string) error
	BuildLogs(context.Context, ConnectionOptions, string, bool) error
	BuildWait(context.Context, ConnectionOptions, string, bool) error
	BuildWatch(context.Context, ConnectionOptions, string, string, bool) error
	BuildRetry(context.Context, ConnectionOptions, string, bool, bool) error
	BuildCancel(context.Context, ConnectionOptions, string) error
}

type ExecutorFactory func(Output) Executor
