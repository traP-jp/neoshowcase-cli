package app

type RestartCommand struct {
	Application string
}

func (RestartCommand) IsCommand() {}

type RestartResult struct {
	Application Application
	State       string
}
