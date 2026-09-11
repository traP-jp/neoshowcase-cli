package build

type GetCommand struct {
	BuildID string
}

func (GetCommand) IsCommand() {}
