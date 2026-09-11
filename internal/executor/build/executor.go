package build

import "github.com/traP-jp/neoshowcase-cli/internal/model"

type Executor struct {
	emit model.Emit
}

func New(emit model.Emit) *Executor {
	return &Executor{emit: emit}
}
