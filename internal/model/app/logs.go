package app

import "time"

type LogsCommand struct {
	Application string
	Follow      bool
	Tail        int32
	Since       time.Time
	Timeout     time.Duration
}

func (LogsCommand) IsCommand() {}
