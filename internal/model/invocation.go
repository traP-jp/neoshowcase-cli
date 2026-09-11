package model

type Invocation struct {
	Command    Command
	Connection Connection
	Output     string
}

type Command interface {
	IsCommand()
}
