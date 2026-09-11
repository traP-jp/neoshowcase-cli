package cli

import (
	"errors"
	"fmt"

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

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

func fail(code int, format string, args ...any) error {
	return &exitError{code: code, err: fmt.Errorf(format, args...)}
}

func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var target *exitError
	if errors.As(err, &target) {
		return target.code
	}
	var modelError *model.Error
	if errors.As(err, &modelError) {
		switch modelError.Kind {
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
