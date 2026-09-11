package build

type GetCommand struct {
	BuildID string
}

func (GetCommand) IsCommand() {}

type GetResult struct {
	Build Build
}

func (GetResult) IsEvent() {}
