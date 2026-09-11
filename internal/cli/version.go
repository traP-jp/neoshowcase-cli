package cli

import "github.com/traP-jp/neoshowcase-cli/internal/model"

type VersionCommand struct{}

func (*VersionCommand) model(version string) model.Command {
	return model.VersionCommand{Version: version}
}
