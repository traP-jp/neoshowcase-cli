package app

import (
	"fmt"
	"strings"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type ListCommand struct{}

func (*ListCommand) Validate(ValidationContext) (model.Command, error) {
	return appmodel.ListCommand{}, nil
}

func RenderList(renderer *clioutput.Renderer, result appmodel.ListResult) error {
	views := make([]applicationView, 0, len(result.Applications))
	var text strings.Builder
	fmt.Fprintln(&text, "ID\tNAME\tCOMMIT\tSTATE")
	for _, application := range result.Applications {
		views = append(views, applicationViewOf(application))
		fmt.Fprintf(&text, "%s\t%s\t%s\t%s\n", application.ID, application.Name, application.Commit, application.ContainerState)
	}
	if renderer.Format() == "jsonl" {
		for _, view := range views {
			if err := renderer.WriteJSONLine(view); err != nil {
				return err
			}
		}
		return nil
	}
	return renderer.WriteValue(views, strings.TrimSuffix(text.String(), "\n"))
}
