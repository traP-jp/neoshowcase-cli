package build

type CancelCommand struct {
	BuildID string
}

func (CancelCommand) IsCommand() {}

type CancelResult struct {
	Build Build
	State string
}
