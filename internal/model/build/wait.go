package build

import "time"

type WaitCommand struct {
	BuildID string
	Logs    bool
	Timeout time.Duration
}

func (WaitCommand) IsCommand() {}
