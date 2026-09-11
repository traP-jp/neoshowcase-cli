package cli

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	defaultTimeout         = 10 * time.Minute
	defaultHistoricalLines = int32(5000)
)

type runtime struct {
	out, errOut io.Writer
	version     string

	endpoint, user, authHeader, configPath string
	output, logLevel                       string
	noColor, insecure, allowMutable        bool
	logger                                 *logrus.Logger
	executor                               Executor
}

type CLI struct {
	command *cobra.Command
	logger  *logrus.Logger
}

func New(out, errOut io.Writer, version string, executorFactory ExecutorFactory) *CLI {
	logger := newLogger(errOut)
	rt := &runtime{out: out, errOut: errOut, version: version, logger: logger}
	rt.executor = executorFactory(rt)
	command := rt.rootCommand()
	command.SetOut(out)
	command.SetErr(errOut)
	return &CLI{command: command, logger: logger}
}

func (c *CLI) Command() *cobra.Command {
	return c.command
}

func (c *CLI) LogError(err error) {
	logCommandError(c.logger, err)
}

func (rt *runtime) rootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "neoshowcase-cli",
		Short: "Third-party operational CLI for NeoShowcase",
		Long: `Third-party operational CLI for NeoShowcase.

Information and monitoring commands support text, JSON, and JSON Lines output.
Streaming commands emit one independently parseable JSON record per line in
either machine-readable mode. Command results are written to stdout and
diagnostics to stderr. Times are emitted as RFC 3339 UTC values.

Every state-changing command requires --allow-mutable-operation on that
invocation. This permission cannot be enabled through an environment variable
or configuration file.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
				rt.noColor = true
			}
			if rt.logger == nil {
				rt.logger = newLogger(rt.errOut)
			}
			if err := configureLogger(rt.logger, rt.logLevel, rt.noColor); err != nil {
				return err
			}
			rt.logger.WithField("command", cmd.CommandPath()).Debug("executing command")
			if cmd.Annotations["mutable"] == "true" && !rt.allowMutable {
				return fail(ExitUsage, "this command changes NeoShowcase state; pass --allow-mutable-operation for this invocation")
			}
			if rt.output != "text" && rt.output != "json" && rt.output != "jsonl" {
				return fail(ExitUsage, "--output must be text, json, or jsonl")
			}
			return nil
		},
	}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fail(ExitUsage, "%v", err)
	})
	flags := root.PersistentFlags()
	flags.StringVar(&rt.endpoint, "endpoint", "", "NeoShowcase gateway URL (NEOSHOWCASE_ENDPOINT)")
	flags.StringVar(&rt.user, "user", "", "NeoShowcase user (NEOSHOWCASE_USER)")
	flags.StringVar(&rt.authHeader, "auth-header", "", "trusted proxy authentication header (NEOSHOWCASE_AUTH_HEADER)")
	flags.StringVar(&rt.configPath, "config", "", "configuration file (JSON or key=value format)")
	flags.StringVarP(&rt.output, "output", "o", "text", "output format: text, json, or jsonl")
	flags.StringVar(&rt.logLevel, "log-level", "warn", "diagnostic log level")
	flags.BoolVar(&rt.noColor, "no-color", false, "disable ANSI colors")
	flags.BoolVar(&rt.insecure, "insecure-skip-verify", false, "DANGER: disable TLS certificate verification")
	flags.BoolVar(&rt.allowMutable, "allow-mutable-operation", false, "explicitly allow one mutable operation")
	root.AddCommand(rt.appCommand(), rt.buildCommand(), rt.versionCommand())
	return root
}

func (rt *runtime) connectionOptions(cmd *cobra.Command) (ConnectionOptions, error) {
	fileConfig, err := loadConfig(rt.configPath, cmd.Flags().Changed("config"))
	if err != nil {
		return ConnectionOptions{}, fail(ExitUsage, "%v", err)
	}
	value := func(flagName, flagValue, envValue, fileValue, fallback string) string {
		if cmd.Flags().Changed(flagName) {
			return strings.TrimSpace(flagValue)
		}
		return firstNonEmpty(envValue, fileValue, fallback)
	}
	options := ConnectionOptions{
		Endpoint:   value("endpoint", rt.endpoint, os.Getenv("NEOSHOWCASE_ENDPOINT"), fileConfig.Endpoint, ""),
		User:       value("user", rt.user, os.Getenv("NEOSHOWCASE_USER"), fileConfig.User, ""),
		AuthHeader: value("auth-header", rt.authHeader, os.Getenv("NEOSHOWCASE_AUTH_HEADER"), fileConfig.AuthHeader, defaultAuthHeader),
		Insecure:   rt.insecure,
	}
	if options.Endpoint == "" {
		return ConnectionOptions{}, fail(ExitUsage, "NeoShowcase endpoint is required (--endpoint or NEOSHOWCASE_ENDPOINT)")
	}
	if options.Insecure {
		rt.logger.Warn("TLS certificate verification is disabled")
	}
	return options, nil
}

func (rt *runtime) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return writeValue(rt.out, rt.output, map[string]string{"version": rt.version}, "neoshowcase-cli "+rt.version)
		},
	}
}

func validateTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fail(ExitUsage, "--timeout must be greater than zero")
	}
	return nil
}

func mutable(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations["mutable"] = "true"
	return cmd
}
