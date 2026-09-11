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

type Log struct {
	ApplicationID string
	Time          time.Time
	Text          string
}

type LogsResult struct {
	Logs      []Log
	Streaming bool
}
