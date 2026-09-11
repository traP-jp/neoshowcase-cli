package model

import "fmt"

type ErrorKind int

const (
	ErrorNotFound ErrorKind = iota + 1
	ErrorFailure
	ErrorTimeout
	ErrorInterrupt
)

type Error struct {
	Kind ErrorKind
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewError(kind ErrorKind, format string, args ...any) error {
	return &Error{Kind: kind, Err: fmt.Errorf(format, args...)}
}
