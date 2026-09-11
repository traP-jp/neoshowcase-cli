package executor

import "github.com/traP-jp/neoshowcase-cli/internal/cli"

type Executor struct {
	output cli.Output
}

func New(output cli.Output) cli.Executor {
	return &Executor{output: output}
}
