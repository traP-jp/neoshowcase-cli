package build

import "time"

type WatchCommand struct {
	Application string
	Commit      string
	Logs        bool
	Timeout     time.Duration
}

func (WatchCommand) IsCommand() {}
