package model

type Event interface {
	IsEvent()
}

type Emit func(Event) error
