package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/spf13/cobra"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
)

func (rt *runtime) buildCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "build", Short: "Inspect and operate builds"}
	cmd.AddCommand(rt.buildListCommand(), rt.buildGetCommand(), rt.buildLogsCommand(), rt.buildWaitCommand(), rt.buildWatchCommand(), rt.buildRetryCommand(), rt.buildCancelCommand())
	return cmd
}

func (rt *runtime) buildListCommand() *cobra.Command {
	var page, limit int32
	cmd := &cobra.Command{
		Use: "list [application]", Short: "List one page of builds", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if page < 0 || limit <= 0 {
				return fail(ExitUsage, "--page must be non-negative and --limit must be positive")
			}
			client, err := rt.client(cmd)
			if err != nil {
				return err
			}
			var builds []*api.Build
			scopedToApplication := len(args) == 1
			names := map[string]string{}
			if len(args) == 1 {
				app, err := resolveApplication(cmd.Context(), client, args[0])
				if err != nil {
					return err
				}
				names[app.GetId()] = app.GetName()
				response, err := client.GetBuilds(cmd.Context(), connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()}))
				if err != nil {
					return rpcError("list application builds", err)
				}
				builds = response.Msg.GetBuilds()
			} else {
				response, err := client.GetAllBuilds(cmd.Context(), connect.NewRequest(&api.GetAllBuildsRequest{Page: page, Limit: limit}))
				if err != nil {
					return rpcError("list builds", err)
				}
				builds = response.Msg.GetBuilds()
				// Keep the command at two O(1) requests while enriching each build
				// without an N+1 GetApplication sequence.
				apps, err := client.GetApplications(cmd.Context(), connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
				if err != nil {
					return rpcError("list applications for build names", err)
				}
				for _, app := range apps.Msg.GetApplications() {
					names[app.GetId()] = app.GetName()
				}
			}
			sort.SliceStable(builds, func(i, j int) bool { return builds[i].GetQueuedAt().AsTime().After(builds[j].GetQueuedAt().AsTime()) })
			if scopedToApplication {
				start := int64(page) * int64(limit)
				if start >= int64(len(builds)) {
					builds = nil
				} else {
					end := start + int64(limit)
					if end > int64(len(builds)) {
						end = int64(len(builds))
					}
					builds = builds[int(start):int(end)]
				}
			}
			views := make([]buildView, 0, len(builds))
			var text strings.Builder
			fmt.Fprintln(&text, "ID\tAPPLICATION\tCOMMIT\tSTATUS")
			for _, build := range builds {
				views = append(views, viewBuild(build, names[build.GetApplicationId()]))
				application := names[build.GetApplicationId()]
				if application == "" {
					application = build.GetApplicationId()
				}
				fmt.Fprintf(&text, "%s\t%s\t%s\t%s\n", build.GetId(), application, build.GetCommit(), build.GetStatus())
			}
			if rt.output == "jsonl" {
				for _, view := range views {
					if err := writeJSONLine(rt.out, view); err != nil {
						return err
					}
				}
				return nil
			}
			return writeValue(rt.out, rt.output, views, strings.TrimSuffix(text.String(), "\n"))
		},
	}
	cmd.Flags().Int32Var(&page, "page", 0, "zero-indexed page")
	cmd.Flags().Int32Var(&limit, "limit", 20, "number of builds in the page")
	return cmd
}

func (rt *runtime) buildGetCommand() *cobra.Command {
	return &cobra.Command{Use: "get <build-id>", Short: "Get one build", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		build, err := getBuild(cmd.Context(), client, args[0])
		if err != nil {
			return err
		}
		view := viewBuild(build, "")
		text := fmt.Sprintf("ID: %s\nApplication ID: %s\nCommit: %s\nStatus: %s\nRetriable: %t", view.ID, view.ApplicationID, view.Commit, view.Status, view.Retriable)
		return writeValue(rt.out, rt.output, view, text)
	}}
}

func getBuild(ctx context.Context, client apiClient, id string) (*api.Build, error) {
	response, err := client.GetBuild(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: id}))
	if err != nil {
		return nil, rpcError("get build", err)
	}
	return response.Msg, nil
}

func terminal(status api.BuildStatus) bool {
	switch status {
	case api.BuildStatus_SUCCEEDED, api.BuildStatus_FAILED, api.BuildStatus_CANCELLED, api.BuildStatus_SKIPPED:
		return true
	default:
		return false
	}
}

func buildResult(build *api.Build) error {
	if build.GetStatus() == api.BuildStatus_SUCCEEDED {
		return nil
	}
	if terminal(build.GetStatus()) {
		return fail(ExitFailure, "build %s finished with status %s", build.GetId(), build.GetStatus())
	}
	return fmt.Errorf("build %s is not in a terminal state (%s)", build.GetId(), build.GetStatus())
}

func waitTick(ctx context.Context) error {
	timer := time.NewTimer(pollInterval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return contextError(ctx, ctx.Err())
	case <-timer.C:
		return nil
	}
}

func (rt *runtime) outputBuildLog(buildID string, data []byte, streaming bool) error {
	if rt.output == "text" {
		_, err := rt.out.Write(data)
		return err
	}
	entry := logView{BuildID: buildID, Log: string(data)}
	if !streaming && rt.output != "jsonl" {
		return writeValue(rt.out, rt.output, entry, "")
	}
	return writeLog(rt.out, rt.output, entry, streaming)
}

func (rt *runtime) streamBuildLog(ctx context.Context, client apiClient, buildID string) error {
	stream, err := client.GetBuildLogStream(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: buildID}))
	if err != nil {
		return contextError(ctx, rpcError("open build log stream", err))
	}
	for stream.Receive() {
		if err := rt.outputBuildLog(buildID, stream.Msg().GetLog(), true); err != nil {
			return err
		}
	}
	if err := stream.Err(); err != nil {
		return contextError(ctx, rpcError("build log stream disconnected", err))
	}
	if ctx.Err() != nil {
		return contextError(ctx, ctx.Err())
	}
	return nil
}

// monitorBuild polls only while a build is queued or while logs are not being
// streamed. A normal stream end is accepted only after one final GetBuild proves
// that the build reached a terminal state.
func (rt *runtime) monitorBuild(ctx context.Context, client apiClient, initial *api.Build, logs bool) (*api.Build, error) {
	build := initial
	for build.GetStatus() == api.BuildStatus_QUEUED {
		if err := waitTick(ctx); err != nil {
			return nil, err
		}
		var err error
		build, err = getBuild(ctx, client, build.GetId())
		if err != nil {
			return nil, contextError(ctx, err)
		}
	}
	if terminal(build.GetStatus()) {
		if logs {
			response, err := client.GetBuildLog(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()}))
			if err != nil {
				return nil, contextError(ctx, rpcError("get build log", err))
			}
			if err := rt.outputBuildLog(build.GetId(), response.Msg.GetLog(), true); err != nil {
				return nil, err
			}
		}
		return build, nil
	}
	if logs {
		if err := rt.streamBuildLog(ctx, client, build.GetId()); err != nil {
			return nil, err
		}
		final, err := getBuild(ctx, client, build.GetId())
		if err != nil {
			return nil, contextError(ctx, err)
		}
		if !terminal(final.GetStatus()) {
			return nil, fmt.Errorf("build log stream ended while build %s remained %s", final.GetId(), final.GetStatus())
		}
		return final, nil
	}
	for !terminal(build.GetStatus()) {
		if err := waitTick(ctx); err != nil {
			return nil, err
		}
		var err error
		build, err = getBuild(ctx, client, build.GetId())
		if err != nil {
			return nil, contextError(ctx, err)
		}
	}
	return build, nil
}

func (rt *runtime) writeBuildResult(build *api.Build, appName string, streaming bool) error {
	view := viewBuild(build, appName)
	if streaming && rt.output == "text" {
		_, err := fmt.Fprintf(rt.errOut, "build %s: %s\n", build.GetId(), build.GetStatus())
		return err
	}
	if streaming && rt.output != "text" {
		return writeJSONLine(rt.out, view)
	}
	return writeValue(rt.out, rt.output, view, fmt.Sprintf("build %s: %s", build.GetId(), build.GetStatus()))
}

func (rt *runtime) buildLogsCommand() *cobra.Command {
	var follow bool
	var timeout time.Duration
	cmd := &cobra.Command{Use: "logs <build-id>", Short: "Print build logs", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateTimeout(timeout); err != nil {
			return err
		}
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		build, err := getBuild(ctx, client, args[0])
		if err != nil {
			return contextError(ctx, err)
		}
		if terminal(build.GetStatus()) {
			response, err := client.GetBuildLog(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()}))
			if err != nil {
				return contextError(ctx, rpcError("get build log", err))
			}
			return rt.outputBuildLog(build.GetId(), response.Msg.GetLog(), false)
		}
		_ = follow // Accepted for compatibility; in-progress build logs always follow the sole server stream.
		final, err := rt.monitorBuild(ctx, client, build, true)
		if err != nil {
			return err
		}
		return buildResult(final)
	}}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "wait for and stream an in-progress build")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout")
	return cmd
}

func (rt *runtime) buildWaitCommand() *cobra.Command {
	var logs bool
	var timeout time.Duration
	cmd := &cobra.Command{Use: "wait <build-id>", Short: "Wait for a build to finish", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateTimeout(timeout); err != nil {
			return err
		}
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		build, err := getBuild(ctx, client, args[0])
		if err != nil {
			return contextError(ctx, err)
		}
		final, err := rt.monitorBuild(ctx, client, build, logs)
		if err != nil {
			return err
		}
		if err := rt.writeBuildResult(final, "", logs); err != nil {
			return err
		}
		return buildResult(final)
	}}
	cmd.Flags().BoolVar(&logs, "logs", false, "print build logs while waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout")
	return cmd
}

func newestMatching(builds []*api.Build, commit string, excluded map[string]struct{}) *api.Build {
	var found *api.Build
	for _, build := range builds {
		if build.GetCommit() != commit {
			continue
		}
		if _, skip := excluded[build.GetId()]; skip {
			continue
		}
		if found == nil || build.GetQueuedAt().AsTime().After(found.GetQueuedAt().AsTime()) {
			found = build
		}
	}
	return found
}

func watchBuild(ctx context.Context, client apiClient, appID, commit string, excluded map[string]struct{}) (*api.Build, error) {
	for {
		response, err := client.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: appID}))
		if err != nil {
			return nil, contextError(ctx, rpcError("watch for build", err))
		}
		if found := newestMatching(response.Msg.GetBuilds(), commit, excluded); found != nil {
			return found, nil
		}
		if err := waitTick(ctx); err != nil {
			return nil, err
		}
	}
}

func (rt *runtime) buildWatchCommand() *cobra.Command {
	var commit string
	var logs bool
	var timeout time.Duration
	cmd := &cobra.Command{Use: "watch <application>", Short: "Watch for a commit's build and wait for it", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(commit) == "" {
			return fail(ExitUsage, "--commit is required")
		}
		if err := validateTimeout(timeout); err != nil {
			return err
		}
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		app, err := resolveApplication(ctx, client, args[0])
		if err != nil {
			return contextError(ctx, err)
		}
		build, err := watchBuild(ctx, client, app.GetId(), commit, nil)
		if err != nil {
			return err
		}
		final, err := rt.monitorBuild(ctx, client, build, logs)
		if err != nil {
			return err
		}
		if err := rt.writeBuildResult(final, app.GetName(), logs); err != nil {
			return err
		}
		return buildResult(final)
	}}
	cmd.Flags().StringVar(&commit, "commit", "", "commit SHA to watch (required)")
	cmd.Flags().BoolVar(&logs, "logs", false, "print build logs while waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout")
	return cmd
}

func buildIDs(builds []*api.Build) map[string]struct{} {
	ids := make(map[string]struct{}, len(builds))
	for _, build := range builds {
		ids[build.GetId()] = struct{}{}
	}
	return ids
}

func (rt *runtime) appRebuildCommand() *cobra.Command {
	var commit string
	var wait, logs bool
	var timeout time.Duration
	cmd := mutable(&cobra.Command{Use: "rebuild <application>", Short: "Rebuild an application commit", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if logs {
			wait = true
		}
		if wait {
			if err := validateTimeout(timeout); err != nil {
				return err
			}
		}
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		cancel := func() {}
		if wait {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		}
		defer cancel()
		app, err := resolveApplication(ctx, client, args[0])
		if err != nil {
			return contextError(ctx, err)
		}
		selectedCommit := strings.TrimSpace(commit)
		if selectedCommit == "" {
			selectedCommit = strings.TrimSpace(app.GetCommit())
		}
		if selectedCommit == "" {
			return fmt.Errorf("application %s has no current commit; no build was started", app.GetId())
		}
		var excluded map[string]struct{}
		if wait {
			before, err := client.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()}))
			if err != nil {
				return contextError(ctx, rpcError("snapshot builds before rebuild", err))
			}
			excluded = buildIDs(before.Msg.GetBuilds())
		}
		_, err = client.RetryCommitBuild(ctx, connect.NewRequest(&api.RetryCommitBuildRequest{ApplicationId: app.GetId(), Commit: selectedCommit}))
		if err != nil {
			return contextError(ctx, rpcError("rebuild application", err))
		}
		if !wait {
			value := map[string]any{"operation": "app.rebuild", "application_id": app.GetId(), "application_name": app.GetName(), "commit": selectedCommit, "state": "requested"}
			return writeValue(rt.out, rt.output, value, fmt.Sprintf("rebuild requested: %s (%s) commit %s", app.GetName(), app.GetId(), selectedCommit))
		}
		build, err := watchBuild(ctx, client, app.GetId(), selectedCommit, excluded)
		if err != nil {
			return err
		}
		final, err := rt.monitorBuild(ctx, client, build, logs)
		if err != nil {
			return err
		}
		if err := rt.writeBuildResult(final, app.GetName(), logs); err != nil {
			return err
		}
		return buildResult(final)
	}})
	cmd.Flags().StringVar(&commit, "commit", "", "commit SHA (defaults to the application's current commit)")
	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the new build to finish")
	cmd.Flags().BoolVar(&logs, "logs", false, "print logs while waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout when waiting")
	return cmd
}

func (rt *runtime) buildRetryCommand() *cobra.Command {
	var wait, logs bool
	var timeout time.Duration
	cmd := mutable(&cobra.Command{Use: "retry <build-id>", Short: "Retry a retriable build", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if logs {
			wait = true
		}
		if wait {
			if err := validateTimeout(timeout); err != nil {
				return err
			}
		}
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		cancel := func() {}
		if wait {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		}
		defer cancel()
		build, err := getBuild(ctx, client, args[0])
		if err != nil {
			return contextError(ctx, err)
		}
		if !build.GetRetriable() {
			return fmt.Errorf("build %s is not retriable; no build was started", build.GetId())
		}
		var excluded map[string]struct{}
		if wait {
			before, err := client.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: build.GetApplicationId()}))
			if err != nil {
				return contextError(ctx, rpcError("snapshot builds before retry", err))
			}
			excluded = buildIDs(before.Msg.GetBuilds())
		}
		_, err = client.RetryCommitBuild(ctx, connect.NewRequest(&api.RetryCommitBuildRequest{ApplicationId: build.GetApplicationId(), Commit: build.GetCommit()}))
		if err != nil {
			return contextError(ctx, rpcError("retry build", err))
		}
		if !wait {
			value := map[string]any{"operation": "build.retry", "source_build_id": build.GetId(), "application_id": build.GetApplicationId(), "commit": build.GetCommit(), "state": "requested"}
			return writeValue(rt.out, rt.output, value, fmt.Sprintf("retry requested: build %s commit %s", build.GetId(), build.GetCommit()))
		}
		newBuild, err := watchBuild(ctx, client, build.GetApplicationId(), build.GetCommit(), excluded)
		if err != nil {
			return err
		}
		final, err := rt.monitorBuild(ctx, client, newBuild, logs)
		if err != nil {
			return err
		}
		if err := rt.writeBuildResult(final, "", logs); err != nil {
			return err
		}
		return buildResult(final)
	}})
	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the new build to finish")
	cmd.Flags().BoolVar(&logs, "logs", false, "print logs while waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout when waiting")
	return cmd
}

func (rt *runtime) buildCancelCommand() *cobra.Command {
	return mutable(&cobra.Command{Use: "cancel <build-id>", Short: "Cancel an in-progress build", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		build, err := getBuild(cmd.Context(), client, args[0])
		if err != nil {
			return err
		}
		if terminal(build.GetStatus()) {
			value := map[string]any{"operation": "build.cancel", "build_id": build.GetId(), "application_id": build.GetApplicationId(), "commit": build.GetCommit(), "status": build.GetStatus().String(), "state": "already terminal; no change"}
			return writeValue(rt.out, rt.output, value, fmt.Sprintf("build %s is already %s; no change", build.GetId(), build.GetStatus()))
		}
		_, err = client.CancelBuild(cmd.Context(), connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()}))
		if err != nil {
			return rpcError("cancel build", err)
		}
		value := map[string]any{"operation": "build.cancel", "build_id": build.GetId(), "application_id": build.GetApplicationId(), "commit": build.GetCommit(), "status": build.GetStatus().String(), "state": "cancel requested"}
		return writeValue(rt.out, rt.output, value, fmt.Sprintf("cancel requested: build %s (%s)", build.GetId(), build.GetStatus()))
	}})
}
