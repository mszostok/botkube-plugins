package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/hashicorp/go-plugin"
	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/executor"
	"github.com/kubeshop/botkube/pkg/pluginx"

	"go.szostok.io/botkube-plugins/internal/cliccando"
	"go.szostok.io/botkube-plugins/internal/getter"
	"go.szostok.io/botkube-plugins/internal/mathx"
	"go.szostok.io/botkube-plugins/internal/osx"
)

const (
	pluginName  = "cliccando"
	description = "Provides an easy way to define interactive messages"
)

// version is set via ldflags by GoReleaser.
var version = "dev"

// Config holds the executor configuration.
type (
	Config struct {
		TemplateSource []getter.Source `yaml:"templateSource"`
		Templates      Templates       `yaml:"templates"`
		TmpDir         osx.TmpDir      `yaml:"tmpDir"`
	}

	Templates []Template

	Template struct {
		Trigger Trigger `yaml:"trigger"`
		Message Message `yaml:"message"`
	}

	Message struct {
		Buttons  api.Buttons `yaml:"buttons"`
		Header   string      `yaml:"header"`
		Paginate Paginate    `yaml:"paginate"`
	}

	Paginate struct {
		Page int `yaml:"page"`
	}
	Trigger struct {
		Command string `yaml:"command"`
	}
)

// CliccandoExecutor implements the Botkube executor plugin interface.
type CliccandoExecutor struct{}

// Metadata returns details about the Echo plugin.
func (CliccandoExecutor) Metadata(context.Context) (api.MetadataOutput, error) {
	return api.MetadataOutput{
		Version:     version,
		Description: description,
		JSONSchema: api.JSONSchema{
			Value: heredoc.Docf(`{
			  "$schema": "http://json-schema.org/draft-04/schema#",
			  "title": "echo",
			  "description": "%s",
			  "type": "object",
			  "properties": {
			    "formatOptions": {
			      "description": "Options to format echoed string",
			      "type": "array",
			      "items": {
			        "type": "string",
			        "enum": [ "bold", "italic" ]
			      }
			    }
			  },
			  "additionalProperties": false
			}`, description),
		},
	}, nil
}

// Execute returns a given command as a response.
//
//nolint:gocritic  //hugeParam: in is heavy (80 bytes); consider passing it by pointer
func (e CliccandoExecutor) Execute(ctx context.Context, in executor.ExecuteInput) (executor.ExecuteOutput, error) {
	var cfg Config
	err := pluginx.MergeExecutorConfigs(in.Configs, &cfg)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	cmd := cliccando.Parse(pluginName, in.Command)

	templates, err := getter.Load[Template](ctx, cfg.TmpDir.GetDirectory(), cfg.TemplateSource)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}
	templates = append(templates, cfg.Templates...)
	msg := e.get(templates, cmd.ToExecute)
	if msg == nil {
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage("command not found", false),
		}, nil
	}

	start := mathx.Max(cmd.PageIndex*msg.Paginate.Page, len(msg.Buttons)-2)
	stop := mathx.Max(start+msg.Paginate.Page, len(msg.Buttons))
	return executor.ExecuteOutput{
		Message: api.Message{
			OnlyVisibleForYou: true,
			ReplaceOriginal:   cmd.PageIndex > 0,
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
					Buttons: e.getPaginationButtons(msg, cmd),
				},
			},
		},
	}, nil
}

func (CliccandoExecutor) get(templates Templates, command string) *Message {
	for _, tpl := range templates {
		fmt.Println(command)
		fmt.Println(tpl.Trigger.Command)
		if !strings.HasPrefix(command, tpl.Trigger.Command) {
			continue
		}

		return &tpl.Message
	}
	return nil
}

func (CliccandoExecutor) Help(context.Context) (api.Message, error) {
	btnBuilder := api.NewMessageButtonBuilder()
	return api.Message{
		Sections: []api.Section{
			{
				Base: api.Base{
					Header:      "Run `echo` commands",
					Description: description,
				},
				Buttons: []api.Button{
					btnBuilder.ForCommandWithDescCmd("Run", "echo 'hello world'"),
				},
			},
		},
	}, nil
}

func (e CliccandoExecutor) getPaginationButtons(msg *Message, cmd cliccando.Command) []api.Button {
	allItems := len(msg.Buttons)
	if allItems <= msg.Paginate.Page {
		return nil
	}

	btnsBuilder := api.NewMessageButtonBuilder()

	var out []api.Button
	if cmd.PageIndex > 0 {
		out = append(out, btnsBuilder.ForCommandWithoutDesc("Prev", fmt.Sprintf("cliccando %s @page:%d", cmd.ToExecute, mathx.DecreaseWithMin(cmd.PageIndex, 0))))
	}

	if cmd.PageIndex*msg.Paginate.Page < allItems-1 {
		out = append(out, btnsBuilder.ForCommandWithoutDesc("Next", fmt.Sprintf("cliccando %s @page:%d", cmd.ToExecute, mathx.IncreaseWithMax(cmd.PageIndex, allItems-1)), api.ButtonStylePrimary))
	}
	return out
}

func main() {
	executor.Serve(map[string]plugin.Plugin{
		pluginName: &executor.Plugin{
			Executor: &CliccandoExecutor{},
		},
	})
}
