package app

type StartCommand struct {
	Application string
}

func (StartCommand) IsCommand() {}

type StartResult struct {
	Application Application
	State       string
}
