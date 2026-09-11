package app

type StopCommand struct {
	Application string
}

func (StopCommand) IsCommand() {}
