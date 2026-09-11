package build

type Command struct {
	List   ListCommand   `cmd:"" help:"List one page of builds"`
	Get    GetCommand    `cmd:"" help:"Get one build"`
	Logs   LogsCommand   `cmd:"" help:"Print build logs"`
	Wait   WaitCommand   `cmd:"" help:"Wait for a build to finish"`
	Watch  WatchCommand  `cmd:"" help:"Watch for a commit's build and wait for it"`
	Retry  RetryCommand  `cmd:"" help:"Retry a retriable build"`
	Cancel CancelCommand `cmd:"" help:"Cancel an in-progress build"`
}

func (Command) Help() string {
	return `Waiting and streaming operations default to a whole-command timeout of 10 minutes. Build detection and state polling run immediately, then every 11 seconds. On rebuild and retry, --logs implies --wait.

A monitored build exits successfully only when its terminal status is SUCCEEDED. FAILED, CANCELLED, and SKIPPED builds return a non-zero status.`
}

type ValidationContext struct {
	AllowMutable bool
}
