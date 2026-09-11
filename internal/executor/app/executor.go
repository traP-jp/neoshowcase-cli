package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	buildexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/build"
)

type Executor struct {
	output *cli.Renderer
	builds *buildexecutor.Executor
}

func New(output *cli.Renderer, builds *buildexecutor.Executor) *Executor {
	return &Executor{output: output, builds: builds}
}
