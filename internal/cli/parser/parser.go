package parser

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	appcli "github.com/traP-jp/neoshowcase-cli/internal/cli/parser/app"
	buildcli "github.com/traP-jp/neoshowcase-cli/internal/cli/parser/build"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

const description = `Third-party operational CLI for NeoShowcase.

Information and monitoring commands support text, JSON, and JSON Lines output. Streaming commands emit one independently parseable JSON record per line in either machine-readable mode. Command results are written to stdout and diagnostics to stderr. Times are emitted as RFC 3339 UTC values.

Every state-changing command requires --allow-mutable-operation on that invocation. This permission cannot be enabled through an environment variable.`

type Command struct {
	Global GlobalOptions `embed:""`

	App     appcli.Command   `cmd:"" help:"Inspect and operate applications"`
	Build   buildcli.Command `cmd:"" help:"Inspect and operate builds"`
	Version VersionCommand   `cmd:"" help:"Print version information"`
}

type Parser struct {
	parser         *kong.Kong
	command        *Command
	version        string
	logger         *logrus.Logger
	initialization error
	exitRequested  bool
	exitCode       int
}

func New(out, errOut io.Writer, version string) *Parser {
	command := &Command{}
	application := &Parser{
		command: command,
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

func (p *Parser) Parse(args []string) (*model.Invocation, error) {
	if p.initialization != nil {
		return nil, fmt.Errorf("initialize CLI: %w", p.initialization)
	}
	p.exitRequested = false
	p.exitCode = 0
	parsed, err := p.parser.Parse(args)
	if p.exitRequested {
		if p.exitCode == 0 {
			return nil, nil
		}
		return nil, model.NewError(model.ErrorUsage, "CLI requested exit status %d", p.exitCode)
	}
	if err != nil {
		return nil, model.NewError(model.ErrorUsage, "%v", err)
	}

	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		p.command.Global.NoColor = true
	}
	if err := configureLogger(p.logger, p.command.Global.LogLevel, p.command.Global.NoColor); err != nil {
		return nil, err
	}
	selected := parsed.Command()
	commandPath := "neoshowcase-cli"
	if selected != "" {
		commandPath += " " + selected
	}
	p.logger.WithField("command", commandPath).Debug("executing command")

	return validateInvocation(p.command, parsed.Selected().Target.Addr().Interface(), p.version, environment{
		User:       os.Getenv("NEOSHOWCASE_USER"),
		AuthHeader: os.Getenv("NEOSHOWCASE_AUTH_HEADER"),
	}, time.Now())
}

func (p *Parser) Warn(message string) {
	p.logger.Warn(message)
}

func (p *Parser) LogError(err error) {
	logCommandError(p.logger, err)
}
