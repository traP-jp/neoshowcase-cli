package cli

import "time"

type AppCommand struct {
	List    AppListCommand    `cmd:"" help:"List applications"`
	Get     AppGetCommand     `cmd:"" help:"Get one application by ID or exact name"`
	Logs    AppLogsCommand    `cmd:"" help:"Print application logs"`
	Start   AppStartCommand   `cmd:"" help:"Start a stopped application"`
	Stop    AppStopCommand    `cmd:"" help:"Stop a running application"`
	Restart AppRestartCommand `cmd:"" help:"Restart a running application"`
	Rebuild AppRebuildCommand `cmd:"" help:"Rebuild an application commit"`
}

func (AppCommand) Help() string {
	return `An <application> argument accepts either an application ID or a unique exact application name. A missing or ambiguous name fails without making changes.

Restart performs a non-atomic state check followed by StartApplication, so a concurrent server-side state change can race the check.`
}

type AppListCommand struct{}

type AppGetCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

type AppLogsCommand struct {
	Application string        `arg:"" help:"Application ID or exact name"`
	Follow      bool          `short:"f" help:"Follow new log records"`
	Tail        int32         `default:"5000" help:"Maximum number of historical lines"`
	Since       string        `help:"Only records since an RFC 3339 time or relative duration"`
	Timeout     time.Duration `default:"10m" help:"Whole-command timeout when following"`
}

type AppStartCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

type AppStopCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

type AppRestartCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

type AppRebuildCommand struct {
	Application string        `arg:"" help:"Application ID or exact name"`
	Commit      string        `help:"Commit SHA (defaults to the application's current commit)"`
	Wait        bool          `help:"Wait for the new build to finish"`
	Logs        bool          `help:"Print logs while waiting"`
	Timeout     time.Duration `default:"10m" help:"Whole-command timeout when waiting"`
}
