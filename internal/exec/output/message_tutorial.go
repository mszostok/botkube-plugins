package output

import (
	"fmt"

	"github.com/kubeshop/botkube/pkg/api"
	"go.uber.org/zap"

	"go.szostok.io/botkube-plugins/internal/exec/template"
	"go.szostok.io/botkube-plugins/internal/mathx"
	"go.szostok.io/botkube-plugins/internal/state"
)

type TutorialWrapper struct {
	log *zap.Logger
}

func NewTutorialWrapper(log *zap.Logger) *TutorialWrapper {
	return &TutorialWrapper{
		log: log,
	}
}

func (p *TutorialWrapper) RenderMessage(cmd, output string, _ *state.Container, msgCtx *template.Template) (api.Message, error) {
	msg := msgCtx.TutorialMessage

	start := mathx.Max(msg.Paginate.CurrentPage*msg.Paginate.Page, len(msg.Buttons)-2)
	stop := mathx.Max(start+msg.Paginate.Page, len(msg.Buttons))
	return api.Message{
		OnlyVisibleForYou: true,
		ReplaceOriginal:   msg.Paginate.CurrentPage > 0,
		Sections: []api.Section{
			{
				Base: api.Base{
					Header: msg.Header,
				},
			},
			{
				Buttons: msg.Buttons[start:stop],
			},
			{
				Buttons: p.getPaginationButtons(msg, msg.Paginate.CurrentPage, cmd),
			},
		},
	}, nil
}

func (p *TutorialWrapper) getPaginationButtons(msg template.TutorialMessage, pageIndex int, cmd string) []api.Button {
	allItems := len(msg.Buttons)
	if allItems <= msg.Paginate.Page {
		return nil
	}

	btnsBuilder := api.NewMessageButtonBuilder()

	var out []api.Button
	if pageIndex > 0 {
		out = append(out, btnsBuilder.ForCommandWithoutDesc("Prev", fmt.Sprintf("x run %s @page:%d", cmd, mathx.DecreaseWithMin(pageIndex, 0))))
	}

	if pageIndex*msg.Paginate.Page < allItems-1 {
		out = append(out, btnsBuilder.ForCommandWithoutDesc("Next", fmt.Sprintf("x run %s @page:%d", cmd, mathx.IncreaseWithMax(pageIndex, allItems-1)), api.ButtonStylePrimary))
	}
	return out
}
