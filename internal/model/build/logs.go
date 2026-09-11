package build

import "time"

type LogsCommand struct {
	BuildID string
	Follow  bool
	Timeout time.Duration
}

func (LogsCommand) IsCommand() {}
