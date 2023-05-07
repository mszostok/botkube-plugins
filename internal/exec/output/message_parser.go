package output

import (
	"fmt"
	"strconv"
	"strings"
	gotemplate "text/template"

	"github.com/huandu/xstrings"
	"github.com/kubeshop/botkube/pkg/api"
	"github.com/sanity-io/litter"
	"go.uber.org/zap"

	"go.szostok.io/botkube-plugins/internal/exec"
	"go.szostok.io/botkube-plugins/internal/exec/parser"
	"go.szostok.io/botkube-plugins/internal/exec/template"
	"go.szostok.io/botkube-plugins/internal/state"
)

type Parser interface {
	TableSpaceSeparated(in string) parser.TableSpaceSeparatedOutput
}

type CommandParser struct {
	parsers map[string]Parser
	log     *zap.Logger
}

func NewTableCommandParser(log *zap.Logger) *CommandParser {
	return &CommandParser{
		log: log,
		parsers: map[string]Parser{
			"space": &parser.TableSpace{},
		},
	}
}

func (p *CommandParser) RenderMessage(cmd, output string, state *state.Container, msgCtx *template.Template) (api.Message, error) {
	msg := msgCtx.ParseMessage
	parserType := strings.TrimPrefix(msgCtx.Type, "parser:table:")
	parser, found := p.parsers[parserType]
	if !found {
		note := fmt.Sprintf("parser %s is not supported", parserType)
		return api.NewPlaintextMessage(note, false), nil
	}

	litter.Dump(state)
	out := parser.TableSpaceSeparated(output)
	if len(out.Lines) == 0 || len(out.Table.Rows) == 0 {
		return noItemsMsg(), nil
	}

	var sections []api.Section

	// dropdowns
	dropdowns, selectedIdx, err := p.renderDropdowns(msg.Selects, out.Table, cmd, state)
	if err != nil {
		return api.Message{}, err
	}
	sections = append(sections, dropdowns)
	// preview
	preview, err := p.renderPreview(msg, out, selectedIdx)
	if err != nil {
		return api.Message{}, err
	}
	sections = append(sections, preview) // todo check header + 1 line at least

	// actions
	actions, err := p.renderActions(msg, out.Table, cmd, selectedIdx)
	if err != nil {
		return api.Message{}, err
	}
	sections = append(sections, actions)

	return api.Message{
		ReplaceOriginal:   state != nil && state.SelectsBlockID != "", // dropdown clicked, let's do the update
		OnlyVisibleForYou: true,
		Sections:          sections,
	}, nil
}

func (p *CommandParser) renderActions(msgCtx template.ParseMessage, table parser.Table, cmd string, idx int) (api.Section, error) {
	if idx >= len(table.Rows) {
		idx = len(table.Rows) - 1
	}
	btnBuilder := api.NewMessageButtonBuilder()
	var actions []api.OptionItem
	for name, tpl := range msgCtx.Actions { // based on the selected item
		out, err := p.renderGoTemplate(tpl, table.Headers, table.Rows[idx])
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

func (p *CommandParser) renderPreview(msgCtx template.ParseMessage, out parser.TableSpaceSeparatedOutput, requestedRow int) (api.Section, error) {
	headerLine := out.Lines[0]

	if requestedRow >= len(out.Table.Rows) {
		requestedRow = len(out.Table.Rows) - 1
	}

	renderLine := p.getPreviewLine(out.Lines, requestedRow)

	preview := fmt.Sprintf("%s\n%s", headerLine, renderLine) // just print the first entry

	if msgCtx.Preview != "" {
		prev, err := p.renderGoTemplate(msgCtx.Preview, out.Table.Headers, out.Table.Rows[requestedRow])
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

func (*CommandParser) getPreviewLine(lines []string, idx int) string {
	if len(lines) < 2 { // exclude the first line for the header
		return ""
	}

	requested := idx + 1
	if len(lines) >= requested {
		return lines[requested]
	}

	return lines[1] // otherwise default first line
}

func (p *CommandParser) renderDropdowns(selects []template.Select, commandData parser.Table, cmd string, state *state.Container) (api.Section, int, error) {

	var (
		dropdowns       []api.Select
		lastSelectedIdx int
	)
	for _, item := range selects {
		var (
			name   = item.Name
			keyTpl = item.KeyTpl
		)
		dropdown, selectedIdx := p.selectDropdown(name, cmd, keyTpl, commandData, state)

		if dropdown != nil {
			dropdowns = append(dropdowns, *dropdown)
			lastSelectedIdx = selectedIdx
		}
	}

	return api.Section{
		Selects: api.Selects{
			ID:    state.GetSelectsBlockID(),
			Items: dropdowns,
		},
	}, lastSelectedIdx, nil
}

func (p *CommandParser) selectDropdown(name, cmd, keyTpl string, table parser.Table, state *state.Container) (*api.Select, int) {
	log := p.log.With(zap.String("selectName", name))
	var options []api.OptionItem
	for idx, row := range table.Rows {
		selectItemName, err := p.renderGoTemplate(keyTpl, table.Headers, row)
		if err != nil {
			return nil, 0
		}
		if selectItemName == "" {
			log.Info("key name is empty for dropdown")
			continue
		}
		options = append(options, api.OptionItem{
			Name:  selectItemName,
			Value: fmt.Sprintf("%s%d", exec.SelectIndexIndicator, idx),
		})
	}

	if len(options) == 0 {
		return nil, 0
	}

	dropdownID := fmt.Sprintf("x run %s", cmd)
	idx := p.resolveSelectIdx(state, dropdownID)
	if idx >= len(options) {
		idx = len(options) - 1
	}

	log.Info("Dropdown rendered", zap.Int("itemsNo", len(options)), zap.Int("selectedItem", idx))
	return &api.Select{
		Type:          api.StaticSelect,
		Name:          name,
		Command:       fmt.Sprintf("%s %s", api.MessageBotNamePlaceholder, dropdownID), // storing select ID under command, so we can easily locate it from a given state
		InitialOption: &options[idx],
		OptionGroups: []api.OptionGroup{
			{
				Name:    name,
				Options: options,
			},
		},
	}, idx
}

func (*CommandParser) resolveSelectIdx(state *state.Container, selectID string) int {
	fmt.Println(selectID)
	item := state.GetField(selectID)
	if item == "" {
		return 0
	}

	fmt.Println(item)
	_, item, _ = strings.Cut(item, exec.SelectIndexIndicator)
	fmt.Println(item)
	val, _ := strconv.Atoi(item)
	fmt.Println(val)
	return val
}

func (*CommandParser) renderGoTemplate(tpl string, cols, rows []string) (string, error) {
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
