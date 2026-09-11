package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func RenderLogs(renderer *core.Writer, result appmodel.LogsResult) error {
	entries := make([]core.Log, 0, len(result.Logs))
	for _, log := range result.Logs {
		entry := core.Log{ApplicationID: log.ApplicationID, Time: utc(log.Time), Text: log.Text}
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
