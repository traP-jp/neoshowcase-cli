package app

import "time"

type Command struct {
	List    ListCommand    `cmd:"" help:"List applications"`
	Get     GetCommand     `cmd:"" help:"Get one application by ID or exact name"`
	Logs    LogsCommand    `cmd:"" help:"Print application logs"`
	Start   StartCommand   `cmd:"" help:"Start a stopped application"`
	Stop    StopCommand    `cmd:"" help:"Stop a running application"`
	Restart RestartCommand `cmd:"" help:"Restart a running application"`
	Rebuild RebuildCommand `cmd:"" help:"Rebuild an application commit"`
}

func (Command) Help() string {
	return `An <application> argument accepts either an application ID or a unique exact application name. A missing or ambiguous name fails without making changes.

Restart performs a non-atomic state check followed by StartApplication, so a concurrent server-side state change can race the check.`
}

type ValidationContext struct {
	AllowMutable bool
	Now          time.Time
}
