package output

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
	"github.com/kubeshop/botkube/pkg/api"

	"go.szostok.io/botkube-plugins/internal/exec"
	"go.szostok.io/botkube-plugins/internal/exec/parser"
)

type InteractiveTable struct{}

func (p InteractiveTable) Parse(cmd string, output string, msgCtx exec.InteractiveItem) (api.Message, error) {
	table, lines := parser.TableSpaceSeparated(output)
	if len(lines) == 0 {
		return noItemsMsg(), nil
	}

	var sections []api.Section

	if len(lines) > 2 {
		dropdowns, err := renderDropdowns(msgCtx, table, cmd, msgCtx.Message.Select.ItemIdx)
		if err != nil {
			return api.Message{}, err
		}
		sections = append(sections, dropdowns)
	}

	// preview
	if len(lines) > 1 { // assumption that the first line is a header
		preview, err := renderPreview(msgCtx, table, lines, msgCtx.Message.Select.ItemIdx)
		if err != nil {
			return api.Message{}, err
		}
		sections = append(sections, preview) // todo check header + 1 line at least
	}

	// actions
	if len(table) > 1 {
		actions, err := renderActions(msgCtx, table, cmd, msgCtx.Message.Select.ItemIdx)
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

func renderActions(msgCtx exec.InteractiveItem, table [][]string, cmd string, idx int) (api.Section, error) {
	headers, firstRow := table[0], table[idx+1]

	btnBuilder := api.NewMessageButtonBuilder()
	var actions []api.OptionItem
	for name, tpl := range msgCtx.Message.Actions { // based on the selected item
		out, err := render(tpl, headers, firstRow)
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
			btnBuilder.ForCommandWithoutDesc("Raw output", fmt.Sprintf("x run %s %s", cmd, exec.NoProcessingIndicator)),
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

func renderPreview(msgCtx exec.InteractiveItem, table [][]string, lines []string, idx int) (api.Section, error) {
	headers, renderRow := table[0], table[1]
	renderLine := lines[1]

	selectedLine := idx + 1
	if len(lines) >= selectedLine {
		renderLine = lines[selectedLine]
	}

	preview := fmt.Sprintf("%s\n%s", lines[0], renderLine) // just print the first entry
	if msgCtx.Message.Preview != "" {
		if len(table) >= selectedLine {
			renderRow = table[selectedLine]
		}

		prev, err := render(msgCtx.Message.Preview, headers, renderRow)
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

func renderDropdowns(msgCtx exec.InteractiveItem, table [][]string, cmd string, idx int) (api.Section, error) {
	headers, rows := table[0], table[1:]

	var dropdowns []api.Select
	parent := api.Select{
		Type:    api.StaticSelect,
		Name:    msgCtx.Message.Select.Name,
		Command: fmt.Sprintf("%s x run %s", api.MessageBotNamePlaceholder, cmd),
	}

	group := api.OptionGroup{
		Name: msgCtx.Message.Select.Name,
	}
	for idx, row := range rows {
		name, err := render(msgCtx.Message.Select.ItemKey, headers, row)
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
			//ID:    "123",
			Items: dropdowns,
		},
	}, nil
}

func render(tpl string, cols []string, rows []string) (string, error) {
	data := map[string]string{}
	for idx, col := range cols {
		col = xstrings.ToCamelCase(strings.ToLower(col))
		data[col] = rows[idx]
	}

	tmpl, err := template.New("tpl").Parse(tpl)
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
