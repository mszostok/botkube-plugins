package output

import (
	"fmt"
	"strings"
	gotemplate "text/template"

	"github.com/huandu/xstrings"
	"github.com/kubeshop/botkube/pkg/api"

	"go.szostok.io/botkube-plugins/internal/exec"
	"go.szostok.io/botkube-plugins/internal/exec/parser"
	"go.szostok.io/botkube-plugins/internal/exec/template"
)

type InteractiveMessage struct{}

func NewInteractiveTableMesage() *InteractiveMessage {
	return &InteractiveMessage{}
}

func (p InteractiveMessage) RenderMessage(cmd, output string, msgCtx *template.Interactive) (api.Message, error) {
	out := parser.TableSpaceSeparated(output)
	if len(out.Lines) == 0 {
		return noItemsMsg(), nil
	}

	var sections []api.Section

	if len(out.Table.Rows) > 0 {
		dropdowns, err := renderDropdowns(msgCtx, out.Table, cmd, msgCtx.Message.Select.ItemIdx)
		if err != nil {
			return api.Message{}, err
		}
		sections = append(sections, dropdowns)
	}

	// preview
	if len(out.Table.Rows) > 0 {
		preview, err := renderPreview(msgCtx, out, msgCtx.Message.Select.ItemIdx)
		if err != nil {
			return api.Message{}, err
		}
		sections = append(sections, preview) // todo check header + 1 line at least
	}

	// actions
	if len(out.Table.Rows) > 0 {
		actions, err := renderActions(msgCtx, out.Table, cmd, msgCtx.Message.Select.ItemIdx)
		if err != nil {
			return api.Message{}, err
		}
		sections = append(sections, actions)
	}

	return api.Message{
		ReplaceOriginal:   msgCtx.Message.Select.Replace,
		OnlyVisibleForYou: true,
		Sections:          sections,
	}, nil
}

func renderActions(msgCtx *template.Interactive, table parser.Table, cmd string, idx int) (api.Section, error) {
	if idx >= len(table.Rows) {
		idx = len(table.Rows) - 1
	}
	btnBuilder := api.NewMessageButtonBuilder()
	var actions []api.OptionItem
	for name, tpl := range msgCtx.Message.Actions { // based on the selected item
		out, err := render(tpl, table.Headers, table.Rows[idx])
		if err != nil {
			return api.Section{}, err
		}
		actions = append(actions, api.OptionItem{
			Name:  name,
			Value: out,
		})
	}
	if len(actions) == 0 {
		return api.Section{}, nil
	}

	return api.Section{
		Buttons: []api.Button{
			btnBuilder.ForCommandWithoutDesc("Raw output", fmt.Sprintf("x run %s %s", cmd, exec.RawOutputIndicator)),
		},
		Selects: api.Selects{
			Items: []api.Select{
				{
					Type:    api.StaticSelect,
					Name:    "Actions",
					Command: fmt.Sprintf("%s x run", api.MessageBotNamePlaceholder),
					OptionGroups: []api.OptionGroup{
						{
							Name:    "Actions",
							Options: actions,
						},
					},
				},
			},
		},
	}, nil
}

func renderPreview(msgCtx *template.Interactive, out parser.TableSpaceSeparatedOutput, requestedRow int) (api.Section, error) {
	headerLine := out.Lines[0]

	if requestedRow >= len(out.Table.Rows) {
		requestedRow = len(out.Table.Rows) - 1
	}

	renderLine := getPreviewLine(out.Lines, requestedRow)

	preview := fmt.Sprintf("%s\n%s", headerLine, renderLine) // just print the first entry

	if msgCtx.Message.Preview != "" {
		prev, err := render(msgCtx.Message.Preview, out.Table.Headers, out.Table.Rows[requestedRow])
		if err != nil {
			return api.Section{}, err
		}
		preview = prev
	}

	return api.Section{
		Base: api.Base{
			Body: api.Body{
				CodeBlock: preview,
			},
		},
	}, nil
}

func getPreviewLine(lines []string, idx int) string {
	if len(lines) < 2 { // exclude the first line for the header
		return ""
	}

	requested := idx + 1
	if len(lines) >= requested {
		return lines[requested]
	}

	return lines[1] // otherwise default first line
}

func renderDropdowns(msgCtx *template.Interactive, table parser.Table, cmd string, idx int) (api.Section, error) {
	var dropdowns []api.Select
	parent := api.Select{
		Type:    api.StaticSelect,
		Name:    msgCtx.Message.Select.Name,
		Command: fmt.Sprintf("%s x run %s", api.MessageBotNamePlaceholder, cmd),
	}

	group := api.OptionGroup{
		Name: msgCtx.Message.Select.Name,
	}
	for idx, row := range table.Rows {
		name, err := render(msgCtx.Message.Select.ItemKey, table.Headers, row)
		if err != nil {
			return api.Section{}, err
		}
		group.Options = append(group.Options, api.OptionItem{
			Name:  name,
			Value: fmt.Sprintf("%s:%d", exec.SelectIndexIndicator, idx),
		})
	}

	if len(group.Options) > 0 {
		parent.InitialOption = &group.Options[idx]
		parent.OptionGroups = []api.OptionGroup{group}
		dropdowns = append(dropdowns, parent)
	}

	return api.Section{
		Selects: api.Selects{
			Items: dropdowns,
		},
	}, nil
}

func render(tpl string, cols, rows []string) (string, error) {
	data := map[string]string{}
	for idx, col := range cols {
		col = xstrings.ToCamelCase(strings.ToLower(col))
		data[col] = rows[idx]
	}

	tmpl, err := gotemplate.New("tpl").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buff strings.Builder
	err = tmpl.Execute(&buff, data)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

func noItemsMsg() api.Message {
	return api.Message{
		Sections: []api.Section{
			{
				Base: api.Base{
					Description: "Not found.",
				},
			},
		},
	}
}
