package app

import (
	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
)

func mutation(operation string, application *api.Application, state string) cli.Mutation {
	return cli.Mutation{Operation: operation, ApplicationID: application.GetId(), ApplicationName: application.GetName(), Commit: application.GetCommit(), State: state}
}
