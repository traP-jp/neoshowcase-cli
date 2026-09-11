package app

type ListCommand struct{}

func (ListCommand) IsCommand() {}

type ListResult struct {
	Applications []Application
}
