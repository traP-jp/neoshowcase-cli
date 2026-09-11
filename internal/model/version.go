package model

type VersionCommand struct {
	Version string
}

func (VersionCommand) IsCommand() {}
