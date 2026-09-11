package cli

import (
	"errors"

	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

const (
	ExitOK        = 0
	ExitFailure   = 1
	ExitUsage     = 2
	ExitNotFound  = 3
	ExitTimeout   = 124
	ExitInterrupt = 130
)

func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var modelError *model.Error
	if errors.As(err, &modelError) {
		switch modelError.Kind {
		case model.ErrorUsage:
			return ExitUsage
		case model.ErrorNotFound:
			return ExitNotFound
		case model.ErrorTimeout:
			return ExitTimeout
		case model.ErrorInterrupt:
			return ExitInterrupt
		default:
			return ExitFailure
		}
	}
	return ExitFailure
}
