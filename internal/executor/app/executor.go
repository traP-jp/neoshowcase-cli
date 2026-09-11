package app

import (
	buildexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/build"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

type Executor struct {
	builds *buildexecutor.Executor
	emit   model.Emit
}

func New(emit model.Emit, builds *buildexecutor.Executor) *Executor {
	return &Executor{builds: builds, emit: emit}
}
