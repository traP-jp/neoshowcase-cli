package build

type CancelCommand struct {
	BuildID string
}

func (CancelCommand) IsCommand() {}
