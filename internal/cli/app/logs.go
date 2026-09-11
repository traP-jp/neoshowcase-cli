package app

import (
	"fmt"
	"time"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type LogsCommand struct {
	Application string        `arg:"" help:"Application ID or exact name"`
	Follow      bool          `short:"f" help:"Follow new log records"`
	Tail        int32         `default:"5000" help:"Maximum number of historical lines"`
	Since       string        `help:"Only records since an RFC 3339 time or relative duration"`
	Timeout     time.Duration `default:"10m" help:"Whole-command timeout when following"`
}

func (command *LogsCommand) Validate(context ValidationContext) (model.Command, error) {
	if command.Tail < 0 {
		return nil, fmt.Errorf("--tail must not be negative")
	}
	if command.Follow {
		if err := validation.Timeout(command.Timeout); err != nil {
			return nil, err
		}
	}
	since, err := parseSince(command.Since, context.Now)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(0)
	if command.Follow {
		timeout = command.Timeout
	}
	return appmodel.LogsCommand{Application: command.Application, Follow: command.Follow, Tail: command.Tail, Since: since, Timeout: timeout}, nil
}

func parseSince(value string, now time.Time) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if duration, err := time.ParseDuration(value); err == nil {
		if duration < 0 {
			return time.Time{}, fmt.Errorf("--since duration must not be negative")
		}
		return now.Add(-duration), nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("--since must be RFC 3339 or a relative duration: %v", err)
	}
	return parsed, nil
}

func RenderLogs(renderer *clioutput.Renderer, result appmodel.LogsResult) error {
	entries := make([]clioutput.Log, 0, len(result.Logs))
	for _, log := range result.Logs {
		entry := clioutput.Log{ApplicationID: log.ApplicationID, Time: utc(log.Time), Text: log.Text}
		entries = append(entries, entry)
		if result.Streaming {
			if err := renderer.WriteLog(entry, true); err != nil {
				return err
			}
		}
	}
	if result.Streaming {
		return nil
	}
	if renderer.Format() == "text" {
		for _, entry := range entries {
			if err := renderer.WriteLog(entry, false); err != nil {
				return err
			}
		}
		return nil
	}
	if renderer.Format() == "jsonl" {
		for _, entry := range entries {
			if err := renderer.WriteJSONLine(entry); err != nil {
				return err
			}
		}
		return nil
	}
	return renderer.WriteValue(entries, "")
}
