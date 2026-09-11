package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
)

func (rt *runtime) appCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "app", Short: "Inspect and operate applications"}
	cmd.AddCommand(
		rt.appListCommand(), rt.appGetCommand(), rt.appLogsCommand(),
		rt.appStartCommand(), rt.appStopCommand(), rt.appRestartCommand(), rt.appRebuildCommand(),
	)
	return cmd
}

func (rt *runtime) appListCommand() *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List applications", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := rt.client(cmd)
			if err != nil {
				return err
			}
			response, err := client.GetApplications(cmd.Context(), connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
			if err != nil {
				return rpcError("list applications", err)
			}
			apps := response.Msg.GetApplications()
			sort.Slice(apps, func(i, j int) bool { return apps[i].GetName() < apps[j].GetName() })
			views := make([]applicationView, 0, len(apps))
			var text strings.Builder
			fmt.Fprintln(&text, "ID\tNAME\tCOMMIT\tSTATE")
			for _, app := range apps {
				views = append(views, viewApplication(app))
				fmt.Fprintf(&text, "%s\t%s\t%s\t%s\n", app.GetId(), app.GetName(), app.GetCommit(), app.GetContainer())
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
}

func (rt *runtime) appGetCommand() *cobra.Command {
	return &cobra.Command{
		Use: "get <application>", Short: "Get one application by ID or exact name", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := rt.client(cmd)
			if err != nil {
				return err
			}
			app, err := resolveApplication(cmd.Context(), client, args[0])
			if err != nil {
				return err
			}
			view := viewApplication(app)
			text := fmt.Sprintf("ID: %s\nName: %s\nCommit: %s\nRunning: %t\nContainer state: %s\nLatest build: %s", view.ID, view.Name, view.Commit, view.Running, view.ContainerState, view.LatestBuildStatus)
			return writeValue(rt.out, rt.output, view, text)
		},
	}
}

func resolveApplication(ctx context.Context, client apiClient, identifier string) (*api.Application, error) {
	response, err := client.GetApplications(ctx, connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
	if err != nil {
		return nil, rpcError("resolve application", err)
	}
	var matches []*api.Application
	for _, app := range response.Msg.GetApplications() {
		if app.GetId() == identifier {
			return app, nil
		}
		if app.GetName() == identifier {
			matches = append(matches, app)
		}
	}
	if len(matches) == 0 {
		return nil, fail(ExitNotFound, "application %q not found", identifier)
	}
	if len(matches) > 1 {
		return nil, fail(ExitNotFound, "application name %q is ambiguous (%d exact matches)", identifier, len(matches))
	}
	return matches[0], nil
}

func parseSince(value string, now time.Time) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if duration, err := time.ParseDuration(value); err == nil {
		if duration < 0 {
			return time.Time{}, fail(ExitUsage, "--since duration must not be negative")
		}
		return now.Add(-duration), nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fail(ExitUsage, "--since must be RFC 3339 or a relative duration: %v", err)
	}
	return parsed, nil
}

func (rt *runtime) appLogsCommand() *cobra.Command {
	var follow bool
	var tail int32
	var sinceRaw string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use: "logs <application>", Short: "Print application logs", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if tail < 0 {
				return fail(ExitUsage, "--tail must not be negative")
			}
			if follow {
				if err := validateTimeout(timeout); err != nil {
					return err
				}
			}
			since, err := parseSince(sinceRaw, time.Now())
			if err != nil {
				return err
			}
			client, err := rt.client(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			cancel := func() {}
			if follow {
				ctx, cancel = context.WithTimeout(ctx, timeout)
			}
			defer cancel()
			app, err := resolveApplication(ctx, client, args[0])
			if err != nil {
				return contextError(ctx, err)
			}
			historyBefore := time.Now().UTC()
			response, err := client.GetOutput(ctx, connect.NewRequest(&api.GetOutputRequest{
				ApplicationId: app.GetId(), Before: timestamppb.New(historyBefore), Limit: tail,
			}))
			if err != nil {
				return contextError(ctx, rpcError("get application logs", err))
			}
			outputs := response.Msg.GetOutputs()
			sort.SliceStable(outputs, func(i, j int) bool { return outputs[i].GetTime().AsTime().Before(outputs[j].GetTime().AsTime()) })
			entries := make([]logView, 0, len(outputs))
			var cursor time.Time
			for _, output := range outputs {
				at := output.GetTime().AsTime()
				if at.After(cursor) {
					cursor = at
				}
				if !since.IsZero() && at.Before(since) {
					continue
				}
				entry := logView{ApplicationID: app.GetId(), Time: utc(output.GetTime()), Log: output.GetLog()}
				entries = append(entries, entry)
				if follow {
					if err := writeLog(rt.out, rt.output, entry, true); err != nil {
						return err
					}
				}
			}
			if !follow {
				if rt.output == "text" {
					for _, entry := range entries {
						if err := writeLog(rt.out, rt.output, entry, false); err != nil {
							return err
						}
					}
					return nil
				}
				if rt.output == "jsonl" {
					for _, entry := range entries {
						if err := writeJSONLine(rt.out, entry); err != nil {
							return err
						}
					}
					return nil
				}
				return writeValue(rt.out, rt.output, entries, "")
			}
			// Start one nanosecond after the last historical entry to prevent the
			// inclusive Loki cursor from duplicating it. If history is empty, begin
			// at command start (or --since when it is later).
			if cursor.IsZero() {
				cursor = historyBefore
			} else {
				cursor = cursor.Add(time.Nanosecond)
			}
			if !since.IsZero() && since.After(cursor) {
				cursor = since
			}
			stream, err := client.GetOutputStream(ctx, connect.NewRequest(&api.GetOutputStreamRequest{ApplicationId: app.GetId(), Begin: timestamppb.New(cursor)}))
			if err != nil {
				return contextError(ctx, rpcError("follow application logs", err))
			}
			for stream.Receive() {
				msg := stream.Msg()
				if err := writeLog(rt.out, rt.output, logView{ApplicationID: app.GetId(), Time: utc(msg.GetTime()), Log: msg.GetLog()}, true); err != nil {
					return err
				}
			}
			if err := stream.Err(); err != nil {
				return contextError(ctx, rpcError("application log stream ended", err))
			}
			if ctx.Err() != nil {
				return contextError(ctx, ctx.Err())
			}
			return fmt.Errorf("application log stream ended unexpectedly")
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow new log records")
	cmd.Flags().Int32Var(&tail, "tail", defaultHistoricalLines, "maximum number of historical lines")
	cmd.Flags().StringVar(&sinceRaw, "since", "", "only records since RFC 3339 time or relative duration")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout when following")
	return cmd
}

func contextError(ctx context.Context, err error) error {
	if ctx.Err() == context.DeadlineExceeded {
		return fail(ExitTimeout, "command timed out")
	}
	if ctx.Err() == context.Canceled {
		return fail(ExitInterrupt, "command interrupted")
	}
	return err
}

func (rt *runtime) appStartCommand() *cobra.Command {
	return mutable(&cobra.Command{Use: "start <application>", Short: "Start a stopped application", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		app, err := resolveApplication(cmd.Context(), client, args[0])
		if err != nil {
			return err
		}
		if app.GetRunning() {
			return rt.writeMutation("app.start", app, "already running; no change")
		}
		_, err = client.StartApplication(cmd.Context(), connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()}))
		if err != nil {
			return rpcError("start application", err)
		}
		return rt.writeMutation("app.start", app, "started")
	}})
}

func (rt *runtime) appStopCommand() *cobra.Command {
	return mutable(&cobra.Command{Use: "stop <application>", Short: "Stop a running application", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		app, err := resolveApplication(cmd.Context(), client, args[0])
		if err != nil {
			return err
		}
		if !app.GetRunning() {
			return rt.writeMutation("app.stop", app, "already stopped; no change")
		}
		_, err = client.StopApplication(cmd.Context(), connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()}))
		if err != nil {
			return rpcError("stop application", err)
		}
		return rt.writeMutation("app.stop", app, "stopped")
	}})
}

func (rt *runtime) appRestartCommand() *cobra.Command {
	return mutable(&cobra.Command{Use: "restart <application>", Short: "Restart a running application", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := rt.client(cmd)
		if err != nil {
			return err
		}
		app, err := resolveApplication(cmd.Context(), client, args[0])
		if err != nil {
			return err
		}
		if !app.GetRunning() {
			return fmt.Errorf("application %q is stopped; refusing to start it (state check is non-atomic)", app.GetId())
		}
		_, err = client.StartApplication(cmd.Context(), connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()}))
		if err != nil {
			return rpcError("restart application (state check is non-atomic)", err)
		}
		return rt.writeMutation("app.restart", app, "restart requested (state check was non-atomic)")
	}})
}

func (rt *runtime) writeMutation(operation string, app *api.Application, message string) error {
	value := map[string]any{"operation": operation, "application_id": app.GetId(), "application_name": app.GetName(), "commit": app.GetCommit(), "state": message}
	return writeValue(rt.out, rt.output, value, fmt.Sprintf("%s: %s (%s): %s", operation, app.GetName(), app.GetId(), message))
}
