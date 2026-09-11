package app

import "time"

type RebuildCommand struct {
	Application string
	Commit      string
	Wait        bool
	Logs        bool
	Timeout     time.Duration
}

func (RebuildCommand) IsCommand() {}
