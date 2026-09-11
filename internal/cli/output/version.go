package output

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func RenderVersion(writer *core.Writer, result model.VersionResult) error {
	return writer.WriteValue(map[string]string{"version": result.Version}, "neoshowcase-cli "+result.Version)
}
