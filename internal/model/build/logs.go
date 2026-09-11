package build

import "time"

type LogsCommand struct {
	BuildID string
	Follow  bool
	Timeout time.Duration
}

func (LogsCommand) IsCommand() {}

type LogResult struct {
	BuildID   string
	Text      string
	Streaming bool
}
