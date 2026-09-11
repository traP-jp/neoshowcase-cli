package cli

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

type GlobalOptions struct {
	Endpoint              string `default:"https://ns.trap.jp" help:"NeoShowcase gateway URL"`
	Output                string `short:"o" default:"text" enum:"text,json,jsonl" help:"Output format"`
	LogLevel              string `name:"log-level" default:"warn" help:"Diagnostic log level"`
	NoColor               bool   `name:"no-color" help:"Disable ANSI colors"`
	InsecureSkipVerify    bool   `name:"insecure-skip-verify" help:"DANGER: disable TLS certificate verification"`
	AllowMutableOperation bool   `name:"allow-mutable-operation" help:"Explicitly allow one mutable operation"`
}

type environment struct {
	User       string
	AuthHeader string
}

func (options GlobalOptions) validateConnection(env environment) (model.Connection, error) {
	connection, err := validation.Connection(options.Endpoint, env.User, env.AuthHeader, options.InsecureSkipVerify)
	if err != nil {
		return model.Connection{}, fail(ExitUsage, "%v", err)
	}
	return connection, nil
}
