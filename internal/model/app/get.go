package app

type GetCommand struct {
	Application string
}

func (GetCommand) IsCommand() {}
