package output

import (
	"github.com/kubeshop/botkube/pkg/api"
	"go.uber.org/zap"

	"go.szostok.io/botkube-plugins/internal/exec/template"
	"go.szostok.io/botkube-plugins/internal/state"
)

type CommandWrapper struct {
	log *zap.Logger
}

func NewCommandWrapper(log *zap.Logger) *CommandWrapper {
	return &CommandWrapper{
		log: log,
	}
}

func (p *CommandWrapper) RenderMessage(_, output string, _ *state.Container, msgCtx *template.Template) (api.Message, error) {
	msg := msgCtx.WrapMessage

	return api.Message{
		Sections: []api.Section{
			{
				Base: api.Base{
					Body: api.Body{
						Plaintext: output,
					},
				},
			},
			{
				Buttons: msg.Buttons,
			},
		},
	}, nil
}
