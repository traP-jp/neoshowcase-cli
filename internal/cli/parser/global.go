package parser

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/parser/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

type GlobalOptions struct {
	Endpoint              string `default:"https://ns.trap.jp" help:"NeoShowcase gateway URL"`
	Output                string `short:"o" default:"text" enum:"text,json,jsonl" help:"Output format"`
	LogLevel              string `name:"log-level" default:"warn" help:"Diagnostic log level"`
	NoColor               bool   `name:"no-color" help:"Disable ANSI colors"`
	AllowMutableOperation bool   `name:"allow-mutable-operation" help:"Explicitly allow one mutable operation"`
}

type environment struct {
	SessionCookie string
}

func (options GlobalOptions) validateConnection(env environment) (model.Connection, error) {
	connection, err := validation.Connection(options.Endpoint, env.SessionCookie)
	if err != nil {
		return model.Connection{}, model.NewError(model.ErrorUsage, "%v", err)
	}
	return connection, nil
}
