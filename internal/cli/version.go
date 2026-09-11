package cli

import (
	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

type VersionCommand struct{}

func (*VersionCommand) model(version string) model.Command {
	return model.VersionCommand{Version: version}
}

func renderVersion(renderer *clioutput.Renderer, result model.VersionResult) error {
	return renderer.WriteValue(map[string]string{"version": result.Version}, "neoshowcase-cli "+result.Version)
}
