package cli

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func (rt *runtime) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Inspect and operate builds",
		Long: `Inspect, monitor, and operate NeoShowcase builds.

Waiting and streaming operations default to a whole-command timeout of 10
minutes. Build detection and state polling run immediately, then every 11
seconds. On rebuild and retry, --logs implies --wait.

A monitored build exits successfully only when its terminal status is
SUCCEEDED. FAILED, CANCELLED, and SKIPPED builds return a non-zero status.`,
	}
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
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			application := ""
			if len(args) == 1 {
				application = args[0]
			}
			return rt.executor.BuildList(cmd.Context(), options, application, page, limit)
		},
	}
	cmd.Flags().Int32Var(&page, "page", 0, "zero-indexed page")
	cmd.Flags().Int32Var(&limit, "limit", 20, "number of builds in the page")
	return cmd
}

func (rt *runtime) buildGetCommand() *cobra.Command {
	return &cobra.Command{
		Use: "get <build-id>", Short: "Get one build", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			return rt.executor.BuildGet(cmd.Context(), options, args[0])
		},
	}
}

func (rt *runtime) buildLogsCommand() *cobra.Command {
	var follow bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use: "logs <build-id>", Short: "Print build logs", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTimeout(timeout); err != nil {
				return err
			}
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			return rt.executor.BuildLogs(ctx, options, args[0], follow)
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "wait for and stream an in-progress build")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout")
	return cmd
}

func (rt *runtime) buildWaitCommand() *cobra.Command {
	var logs bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use: "wait <build-id>", Short: "Wait for a build to finish", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTimeout(timeout); err != nil {
				return err
			}
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			return rt.executor.BuildWait(ctx, options, args[0], logs)
		},
	}
	cmd.Flags().BoolVar(&logs, "logs", false, "print build logs while waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout")
	return cmd
}

func (rt *runtime) buildWatchCommand() *cobra.Command {
	var commit string
	var logs bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use: "watch <application>", Short: "Watch for a commit's build and wait for it", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(commit) == "" {
				return fail(ExitUsage, "--commit is required")
			}
			if err := validateTimeout(timeout); err != nil {
				return err
			}
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			return rt.executor.BuildWatch(ctx, options, args[0], commit, logs)
		},
	}
	cmd.Flags().StringVar(&commit, "commit", "", "commit SHA to watch (required)")
	cmd.Flags().BoolVar(&logs, "logs", false, "print build logs while waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout")
	return cmd
}

func (rt *runtime) buildRetryCommand() *cobra.Command {
	var wait, logs bool
	var timeout time.Duration
	cmd := mutable(&cobra.Command{
		Use: "retry <build-id>", Short: "Retry a retriable build", Args: cobra.ExactArgs(1),
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
			return rt.executor.BuildRetry(ctx, options, args[0], wait, logs)
		},
	})
	cmd.Flags().BoolVar(&wait, "wait", false, "wait for the new build to finish")
	cmd.Flags().BoolVar(&logs, "logs", false, "print logs while waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "whole-command timeout when waiting")
	return cmd
}

func (rt *runtime) buildCancelCommand() *cobra.Command {
	return mutable(&cobra.Command{
		Use: "cancel <build-id>", Short: "Cancel an in-progress build", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := rt.connectionOptions(cmd)
			if err != nil {
				return err
			}
			return rt.executor.BuildCancel(cmd.Context(), options, args[0])
		},
	})
}
