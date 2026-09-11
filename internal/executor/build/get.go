package build

import (
	"context"

	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func (e *Executor) Get(ctx context.Context, options model.Connection, id string) error {
	build, err := get(ctx, client.New(options), id)
	if err != nil {
		return err
	}
	return e.emit(buildmodel.GetResult{Build: ToModel(build, "")})
}
