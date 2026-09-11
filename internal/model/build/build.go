package build

import "time"

type Build struct {
	ID              string
	ApplicationID   string
	ApplicationName string
	Commit          string
	Status          string
	Retriable       bool
	QueuedAt        time.Time
	StartedAt       *time.Time
	UpdatedAt       *time.Time
	FinishedAt      *time.Time
}

type CompletionResult struct {
	Build     Build
	Streaming bool
}

func (CompletionResult) IsEvent() {}
