package app

import (
	"context"

	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func (e *Executor) Get(ctx context.Context, options model.Connection, identifier string) (appmodel.GetResult, error) {
	application, err := client.ResolveApplication(ctx, client.New(options), identifier)
	if err != nil {
		return appmodel.GetResult{}, err
	}
	return appmodel.GetResult{Application: toModel(application)}, nil
}
