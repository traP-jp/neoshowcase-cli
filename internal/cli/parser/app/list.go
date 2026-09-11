package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type ListCommand struct{}

func (*ListCommand) Validate(ValidationContext) (model.Command, error) {
	return appmodel.ListCommand{}, nil
}
