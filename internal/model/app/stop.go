package app

type StopCommand struct {
	Application string
}

func (StopCommand) IsCommand() {}

type StopResult struct {
	Application Application
	State       string
}

func (StopResult) IsEvent() {}
