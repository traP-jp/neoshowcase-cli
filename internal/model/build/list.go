package build

type ListCommand struct {
	Application string
	Page        int32
	Limit       int32
}

func (ListCommand) IsCommand() {}
