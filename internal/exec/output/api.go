package output

import (
	"github.com/kubeshop/botkube/pkg/api"

	"go.szostok.io/botkube-plugins/internal/exec"
)

// Parser describes API for command output parsers.
type Parser interface {
	Parse(executedCmd string, output string, msgCtx exec.Interactive) (api.Message, error)
}
