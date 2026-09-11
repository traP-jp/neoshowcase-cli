package cli

import (
	"fmt"
	"time"

	appcli "github.com/traP-jp/neoshowcase-cli/internal/cli/app"
	buildcli "github.com/traP-jp/neoshowcase-cli/internal/cli/build"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

type appCommandValidator interface {
	Validate(appcli.ValidationContext) (model.Command, error)
}

type buildCommandValidator interface {
	Validate(buildcli.ValidationContext) (model.Command, error)
}

func validateInvocation(command *Command, selected any, version string, env environment, now time.Time) (*model.Invocation, error) {
	var validated model.Command
	requiresConnection := true
	var err error
	switch value := selected.(type) {
	case *VersionCommand:
		validated = value.model(version)
		requiresConnection = false
	case appCommandValidator:
		validated, err = value.Validate(appcli.ValidationContext{AllowMutable: command.Global.AllowMutableOperation, Now: now})
	case buildCommandValidator:
		validated, err = value.Validate(buildcli.ValidationContext{AllowMutable: command.Global.AllowMutableOperation})
	default:
		return nil, fmt.Errorf("unsupported command type %T", selected)
	}
	if err != nil {
		return nil, fail(ExitUsage, "%v", err)
	}
	invocation := &model.Invocation{Command: validated, Output: command.Global.Output}
	if requiresConnection {
		connection, err := command.Global.validateConnection(env)
		if err != nil {
			return nil, err
		}
		invocation.Connection = connection
	}
	return invocation, nil
}
