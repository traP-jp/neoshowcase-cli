package app

import "time"

type Application struct {
	ID                string
	Name              string
	Commit            string
	Running           bool
	ContainerState    string
	LatestBuildStatus string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
