package build

import "github.com/traP-jp/neoshowcase-cli/internal/cli"

type Executor struct {
	output *cli.Renderer
}

func New(output *cli.Renderer) *Executor {
	return &Executor{output: output}
}
