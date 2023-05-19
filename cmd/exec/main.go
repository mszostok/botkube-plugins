package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/alexflint/go-arg"
	"github.com/hashicorp/go-plugin"
	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/executor"
	"github.com/kubeshop/botkube/pkg/pluginx"
	"go.uber.org/zap"

	x "go.szostok.io/botkube-plugins/internal/exec"
	"go.szostok.io/botkube-plugins/internal/exec/output"
	"go.szostok.io/botkube-plugins/internal/exec/template"
	"go.szostok.io/botkube-plugins/internal/formatx"
	"go.szostok.io/botkube-plugins/internal/loggerx"
	"go.szostok.io/botkube-plugins/internal/state"
)

// version is set via ldflags by GoReleaser.
var version = "dev"

const pluginName = "x"

// XExecutor implements Botkube executor plugin.
type XExecutor struct {
	log *zap.Logger
}

func (i *XExecutor) Help(_ context.Context) (api.Message, error) {
	help := heredoc.Doc(`
		Usage:
		  x run [COMMAND] [FLAGS]    Run a specified command with optional flags
		  x install [SOURCE]         Install a binary using the https://github.com/zyedidia/eget syntax.
		
		Usage Examples:
		  # Install the Helm CLI

		  x install https://get.helm.sh/helm-v3.10.3-linux-amd64.tar.gz --file helm    
		  
		  # Run the 'helm list -A' command.

		  x run helm list -A    
		
		Options:
		  -h, --help                 Show this help message`)
	return api.NewCodeBlockMessage(help, true), nil
}

// Metadata returns details about Echo plugin.
func (*XExecutor) Metadata(context.Context) (api.MetadataOutput, error) {
	return api.MetadataOutput{
		Version:      version,
		Description:  "Install and run CLIs directly from chat window without hassle. All magic included.",
		Dependencies: x.GetPluginDependencies(),
		JSONSchema:   jsonSchema(),
	}, nil
}

type (
	Commands struct {
		Install *InstallCmd `arg:"subcommand:install"`
		Run     *RunCmd     `arg:"subcommand:run"`
	}
	InstallCmd struct {
		Tool []string `arg:"positional"`
	}
	RunCmd struct {
		Tool []string `arg:"positional"`
	}
)

func escapePositionals(in string) string {
	for _, name := range []string{"run", "install"} {
		if strings.Contains(in, name) {
			return strings.Replace(in, name, fmt.Sprintf("%s -- ", name), 1)
		}
	}
	return in
}

// Execute returns a given command as response.
//
//nolint:gocritic // hugeParam: in is heavy (80 bytes); consider passing it by pointer
func (i *XExecutor) Execute(ctx context.Context, in executor.ExecuteInput) (executor.ExecuteOutput, error) {
	var cmd Commands
	in.Command = escapePositionals(in.Command)
	err := pluginx.ParseCommand(pluginName, in.Command, &cmd)
	switch err {
	case nil:
	case arg.ErrHelp:
		msg, _ := i.Help(ctx)
		return executor.ExecuteOutput{
			Message: msg,
		}, nil
	default:
		return executor.ExecuteOutput{}, fmt.Errorf("while parsing input command: %w", err)
	}

	var cfg x.Config
	if err := pluginx.MergeExecutorConfigs(in.Configs, &cfg); err != nil {
		return executor.ExecuteOutput{}, err
	}

	renderer := x.NewRenderer()
	err = renderer.Register("parser:table:.*", output.NewTableCommandParser(i.log))
	if err != nil {
		return executor.ExecuteOutput{}, err
	}
	err = renderer.Register("wrapper", output.NewCommandWrapper(i.log))
	if err != nil {
		return executor.ExecuteOutput{}, err
	}
	err = renderer.Register("tutorial", output.NewTutorialWrapper(i.log))
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	switch {
	case cmd.Run != nil:
		tool := formatx.Normalize(strings.Join(cmd.Run.Tool, " "))
		i.log.Info("Running command...", zap.String("tool", tool))

		//
		//err = renderer.Register("builder", output.NewInteractiveBuilderMesage())
		//if err != nil {
		//	return executor.ExecuteOutput{}, err
		//}

		state := state.ExtractSlackState(in.Context.SlackState)
		return x.NewRunner(i.log, renderer).Run(ctx, cfg, state, tool)
	case cmd.Install != nil:
		var (
			tool          = formatx.Normalize(strings.Join(cmd.Install.Tool, " "))
			dir, isCustom = cfg.TmpDir.Get()
			downloadCmd   = fmt.Sprintf("eget %s", tool)
		)

		i.log.Info("Installing binary...", zap.String("dir", dir), zap.Bool("isCustom", isCustom), zap.String("downloadCmd", downloadCmd))
		if _, err := pluginx.ExecuteCommandWithEnvs(ctx, downloadCmd, map[string]string{
			"EGET_BIN": dir,
		}); err != nil {
			return executor.ExecuteOutput{}, err
		}

		templates, err := template.Load(ctx, cfg.TmpDir.GetDirectory(), cfg.Templates)
		if err != nil {
			return executor.ExecuteOutput{}, err
		}

		cmdTemplate, found := templates.FindWithPrefix(fmt.Sprintf("x install %s", tool))
		if !found {
			i.log.Info("Templates config not found for install command")
			return executor.ExecuteOutput{
				Message: api.NewPlaintextMessage("Binary was installed successfully", false),
			}, nil
		}

		render, err := renderer.Get(cmdTemplate.Type) // Message.Type
		if err != nil {
			return executor.ExecuteOutput{}, err
		}
		message, err := render.RenderMessage(tool, "Binary was installed successfully :tada:", nil, &cmdTemplate)
		if err != nil {
			return executor.ExecuteOutput{}, err
		}
		return executor.ExecuteOutput{
			Message: message,
		}, nil

	}
	return executor.ExecuteOutput{
		Message: api.NewPlaintextMessage("Command not supported", false),
	}, nil
}

func main() {
	logger := loggerx.MustNewLogger()
	defer func() {
		_ = logger.Sync()
	}()

	executor.Serve(map[string]plugin.Plugin{
		pluginName: &executor.Plugin{
			Executor: &XExecutor{
				log: logger,
			},
		},
	})
}

// jsonSchema returns JSON schema for the executor.
func jsonSchema() api.JSONSchema {
	return api.JSONSchema{
		Value: heredoc.Docf(`{

			  "$schema": "http://json-schema.org/draft-07/schema#",
			  "title": "x",
			  "description": "Install and run CLIs directly from chat window without hassle. All magic included.",
			  "type": "object",
			  "properties": {
    "templates": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "ref": {
            "type": "string",
            "default": "github.com/mszostok/botkube-plugins//x-templates?ref=hackathon"
          }
        },
        "required": ["ref"],
        "additionalProperties": false
      }
    }
  },
  "required": ["templates"]
			}`),
	}
}
