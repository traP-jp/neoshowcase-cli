package app

type StartCommand struct {
	Application string
}

func (StartCommand) IsCommand() {}
