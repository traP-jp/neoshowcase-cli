package cli

import "time"

type BuildCommand struct {
	List   BuildListCommand   `cmd:"" help:"List one page of builds"`
	Get    BuildGetCommand    `cmd:"" help:"Get one build"`
	Logs   BuildLogsCommand   `cmd:"" help:"Print build logs"`
	Wait   BuildWaitCommand   `cmd:"" help:"Wait for a build to finish"`
	Watch  BuildWatchCommand  `cmd:"" help:"Watch for a commit's build and wait for it"`
	Retry  BuildRetryCommand  `cmd:"" help:"Retry a retriable build"`
	Cancel BuildCancelCommand `cmd:"" help:"Cancel an in-progress build"`
}

func (BuildCommand) Help() string {
	return `Waiting and streaming operations default to a whole-command timeout of 10 minutes. Build detection and state polling run immediately, then every 11 seconds. On rebuild and retry, --logs implies --wait.

A monitored build exits successfully only when its terminal status is SUCCEEDED. FAILED, CANCELLED, and SKIPPED builds return a non-zero status.`
}

type BuildListCommand struct {
	Application string `arg:"" optional:"" help:"Application ID or exact name"`
	Page        int32  `default:"0" help:"Zero-indexed page"`
	Limit       int32  `default:"20" help:"Number of builds in the page"`
}

type BuildGetCommand struct {
	BuildID string `arg:"" name:"build-id" help:"Build ID"`
}

type BuildLogsCommand struct {
	BuildID string        `arg:"" name:"build-id" help:"Build ID"`
	Follow  bool          `short:"f" help:"Wait for and stream an in-progress build"`
	Timeout time.Duration `default:"10m" help:"Whole-command timeout"`
}

type BuildWaitCommand struct {
	BuildID string        `arg:"" name:"build-id" help:"Build ID"`
	Logs    bool          `help:"Print build logs while waiting"`
	Timeout time.Duration `default:"10m" help:"Whole-command timeout"`
}

type BuildWatchCommand struct {
	Application string        `arg:"" help:"Application ID or exact name"`
	Commit      string        `required:"" help:"Commit SHA to watch"`
	Logs        bool          `help:"Print build logs while waiting"`
	Timeout     time.Duration `default:"10m" help:"Whole-command timeout"`
}

type BuildRetryCommand struct {
	BuildID string        `arg:"" name:"build-id" help:"Build ID"`
	Wait    bool          `help:"Wait for the new build to finish"`
	Logs    bool          `help:"Print logs while waiting"`
	Timeout time.Duration `default:"10m" help:"Whole-command timeout when waiting"`
}

type BuildCancelCommand struct {
	BuildID string `arg:"" name:"build-id" help:"Build ID"`
}
