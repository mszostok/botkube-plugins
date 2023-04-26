package output

import (
	api "github.com/kubeshop/botkube/pkg/api"

	"go.szostok.io/botkube-plugins/internal/exec"
)

func BuildMessage(cmd, output string, msgCtx exec.InteractiveItem) (api.Message, error) {
	var parser Parser
	switch msgCtx.Command.Parser {
	case "table":
		parser = InteractiveTable{}
	default:
		return api.Message{
			BaseBody: api.Body{
				Plaintext: "not supported output parser",
			},
		}, nil
	}

	return parser.Parse(cmd, output, msgCtx)
}
