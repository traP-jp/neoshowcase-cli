package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
)

const description = `Third-party operational CLI for NeoShowcase.

Information and monitoring commands support text, JSON, and JSON Lines output. Streaming commands emit one independently parseable JSON record per line in either machine-readable mode. Command results are written to stdout and diagnostics to stderr. Times are emitted as RFC 3339 UTC values.

Every state-changing command requires --allow-mutable-operation on that invocation. This permission cannot be enabled through an environment variable.`

type GlobalOptions struct {
	Endpoint              string `default:"https://ns.trap.jp" help:"NeoShowcase gateway URL"`
	Output                string `short:"o" default:"text" enum:"text,json,jsonl" help:"Output format"`
	LogLevel              string `name:"log-level" default:"warn" help:"Diagnostic log level"`
	NoColor               bool   `name:"no-color" help:"Disable ANSI colors"`
	InsecureSkipVerify    bool   `name:"insecure-skip-verify" help:"DANGER: disable TLS certificate verification"`
	AllowMutableOperation bool   `name:"allow-mutable-operation" help:"Explicitly allow one mutable operation"`
}

type Command struct {
	Global GlobalOptions `embed:""`

	App     AppCommand     `cmd:"" help:"Inspect and operate applications"`
	Build   BuildCommand   `cmd:"" help:"Inspect and operate builds"`
	Version VersionCommand `cmd:"" help:"Print version information"`
}

type VersionCommand struct{}

type Environment struct {
	User       string
	AuthHeader string
}

type Invocation struct {
	Command     *Command
	Selected    any
	CommandName string
	Version     string
	Environment Environment
}

type CLI struct {
	parser         *kong.Kong
	command        *Command
	out, errOut    io.Writer
	version        string
	logger         *logrus.Logger
	initialization error
	exitRequested  bool
	exitCode       int
}

func New(out, errOut io.Writer, version string) *CLI {
	command := &Command{}
	application := &CLI{
		command: command,
		out:     out,
		errOut:  errOut,
		version: version,
		logger:  newLogger(errOut),
	}
	application.parser, application.initialization = kong.New(
		command,
		kong.Name("neoshowcase-cli"),
		kong.Description(description),
		kong.Writers(out, errOut),
		kong.Exit(func(code int) {
			application.exitRequested = true
			application.exitCode = code
		}),
		kong.ConfigureHelp(kong.HelpOptions{NoExpandSubcommands: true}),
	)
	return application
}

func (c *CLI) Parse(args []string) (*Invocation, error) {
	if c.initialization != nil {
		return nil, fmt.Errorf("initialize CLI: %w", c.initialization)
	}
	c.exitRequested = false
	c.exitCode = ExitOK
	parsed, err := c.parser.Parse(args)
	if c.exitRequested {
		if c.exitCode == ExitOK {
			return nil, nil
		}
		return nil, fail(c.exitCode, "CLI requested exit status %d", c.exitCode)
	}
	if err != nil {
		return nil, fail(ExitUsage, "%v", err)
	}

	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		c.command.Global.NoColor = true
	}
	if err := configureLogger(c.logger, c.command.Global.LogLevel, c.command.Global.NoColor); err != nil {
		return nil, err
	}
	selected := parsed.Command()
	commandPath := "neoshowcase-cli"
	if selected != "" {
		commandPath += " " + selected
	}
	c.logger.WithField("command", commandPath).Debug("executing command")

	return &Invocation{
		Command:     c.command,
		Selected:    parsed.Selected().Target.Addr().Interface(),
		CommandName: selected,
		Version:     c.version,
		Environment: Environment{
			User:       os.Getenv("NEOSHOWCASE_USER"),
			AuthHeader: os.Getenv("NEOSHOWCASE_AUTH_HEADER"),
		},
	}, nil
}

func (c *CLI) Renderer(format string) *Renderer {
	return &Renderer{out: c.out, errOut: c.errOut, format: format, logger: c.logger}
}

func (c *CLI) LogError(err error) {
	logCommandError(c.logger, err)
}
