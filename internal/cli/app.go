package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

func (rt *runtime) appCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Inspect and operate applications",
		Long: `Inspect and operate NeoShowcase applications.

An <application> argument accepts either an application ID or a unique exact
application name. A missing or ambiguous name fails without making changes.

Restart performs a non-atomic state check followed by StartApplication, so a
concurrent server-side state change can race the check.`,
	}
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
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			return rt.executor.AppList(cmd.Context(), options)
		},
	}
}

func (rt *runtime) appGetCommand() *cobra.Command {
	return &cobra.Command{
		Use: "get <application>", Short: "Get one application by ID or exact name", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			return rt.executor.AppGet(cmd.Context(), options, args[0])
		},
	}
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
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			cancel := func() {}
			if follow {
				ctx, cancel = context.WithTimeout(ctx, timeout)
			}
			defer cancel()
			return rt.executor.AppLogs(ctx, options, args[0], follow, tail, since)
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow new log records")
	cmd.Flags().Int32Var(&tail, "tail", defaultHistoricalLines, "maximum number of historical lines")
	cmd.Flags().StringVar(&sinceRaw, "since", "", "only records since RFC 3339 time or relative duration")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout when following")
	return cmd
}

func (rt *runtime) appStartCommand() *cobra.Command {
	return mutable(&cobra.Command{
		Use: "start <application>", Short: "Start a stopped application", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			return rt.executor.AppStart(cmd.Context(), options, args[0])
		},
	})
}

func (rt *runtime) appStopCommand() *cobra.Command {
	return mutable(&cobra.Command{
		Use: "stop <application>", Short: "Stop a running application", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			return rt.executor.AppStop(cmd.Context(), options, args[0])
		},
	})
}

func (rt *runtime) appRestartCommand() *cobra.Command {
	return mutable(&cobra.Command{
		Use: "restart <application>", Short: "Restart a running application", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			return rt.executor.AppRestart(cmd.Context(), options, args[0])
		},
	})
}

func (rt *runtime) appRebuildCommand() *cobra.Command {
	var commit string
	var wait, logs bool
	var timeout time.Duration
	cmd := mutable(&cobra.Command{
		Use: "rebuild <application>", Short: "Rebuild an application commit", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if logs {
				wait = true
			}
			if wait {
				if err := validateTimeout(timeout); err != nil {
					return err
				}
			}
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			cancel := func() {}
			if wait {
				ctx, cancel = context.WithTimeout(ctx, timeout)
			}
			defer cancel()
			return rt.executor.AppRebuild(ctx, options, args[0], commit, wait, logs)
		},
	})
	cmd.Flags().StringVar(&commit, "commit", "", "commit SHA (defaults to the application's current commit)")
	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the new build to finish")
	cmd.Flags().BoolVar(&logs, "logs", false, "print logs while waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout when waiting")
	return cmd
}
