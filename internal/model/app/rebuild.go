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

type RebuildRequested struct {
	Application Application
	Commit      string
}

func (RebuildRequested) IsEvent() {}
