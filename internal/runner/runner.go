package runner

import (
	"context"
	"errors"
	"io"

	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	"github.com/traP-jp/neoshowcase-cli/internal/executor"
)

func Execute(ctx context.Context, args []string, out, errOut io.Writer, version string) int {
	application := cli.New(out, errOut, version, executor.New)
	command := application.Command()
	command.SetArgs(args)

	err := command.ExecuteContext(ctx)
	if err == nil {
		return cli.ExitOK
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		application.LogError(errors.New("command timed out"))
		return cli.ExitTimeout
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		application.LogError(errors.New("command interrupted"))
		return cli.ExitInterrupt
	}
	application.LogError(err)
	return cli.ExitCode(err)
}
