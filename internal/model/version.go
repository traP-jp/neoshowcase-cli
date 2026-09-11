package model

type VersionCommand struct {
	Version string
}

func (VersionCommand) IsCommand() {}

type VersionResult struct {
	Version string
}

func (VersionResult) IsEvent() {}
