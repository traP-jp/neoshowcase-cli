package app

type RestartCommand struct {
	Application string
}

func (RestartCommand) IsCommand() {}
