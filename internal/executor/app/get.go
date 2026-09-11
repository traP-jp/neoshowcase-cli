package app

import (
	"context"

	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func (e *Executor) Get(ctx context.Context, options model.Connection, identifier string) error {
	application, err := client.ResolveApplication(ctx, client.New(options), identifier)
	if err != nil {
		return err
	}
	return e.emit(appmodel.GetResult{Application: toModel(application)})
}
