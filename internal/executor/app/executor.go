package app

import (
	buildexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/build"
)

type Executor struct {
	builds *buildexecutor.Executor
}

func New(builds *buildexecutor.Executor) *Executor {
	return &Executor{builds: builds}
}
