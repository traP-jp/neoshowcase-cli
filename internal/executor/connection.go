package executor

import (
	"strings"

	"github.com/traP-jp/neoshowcase-cli/internal/cli"
)

const defaultAuthHeader = "X-Showcase-User"

type ConnectionOptions struct {
	Endpoint   string
	User       string
	AuthHeader string
	Insecure   bool
}

func (e *Executor) connectionOptions(invocation *cli.Invocation) (ConnectionOptions, error) {
	options := ConnectionOptions{
		Endpoint:   strings.TrimSpace(invocation.Command.Global.Endpoint),
		User:       strings.TrimSpace(invocation.Environment.User),
		AuthHeader: firstNonEmpty(invocation.Environment.AuthHeader, defaultAuthHeader),
		Insecure:   invocation.Command.Global.InsecureSkipVerify,
	}
	if options.Insecure {
		e.output.Warn("TLS certificate verification is disabled")
	}
	return options, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
