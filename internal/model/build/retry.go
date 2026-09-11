package build

import "time"

type RetryCommand struct {
	BuildID string
	Wait    bool
	Logs    bool
	Timeout time.Duration
}

func (RetryCommand) IsCommand() {}

type RetryRequested struct {
	Build Build
}

func (RetryRequested) IsEvent() {}
