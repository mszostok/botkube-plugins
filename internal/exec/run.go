package exec

import (
	"context"
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/executor"
	"github.com/kubeshop/botkube/pkg/pluginx"
	"go.uber.org/zap"

	"go.szostok.io/botkube-plugins/internal/exec/template"
	"go.szostok.io/botkube-plugins/internal/osx"
)

type Runner struct {
	log      *zap.Logger
	renderer *Renderer
}

func NewRunner(log *zap.Logger, renderer *Renderer) *Runner {
	return &Runner{
		log:      log,
		renderer: renderer,
	}
}

func (i *Runner) Run(ctx context.Context, cfg Config, tool string) (executor.ExecuteOutput, error) {
	cmd := Parse(tool)
	out, err := runCmd(ctx, cfg.TmpDir, cmd.ToExecute)
	if err != nil {
		i.log.Error("failed to run command", zap.String("command", cmd.ToExecute), zap.Error(err))
		return executor.ExecuteOutput{}, err
	}

	if cmd.IsRawRequired {
		i.log.Info("Raw output was explicitly requested")
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage(out, true),
		}, nil
	}

	templates, err := template.Load(ctx, cfg.TmpDir.GetDirectory(), cfg.Templates)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	cmdTemplate, found := templates.FindWithPrefix(cmd.ToExecute)
	if !found {
		i.log.Info("Interactive config not found for command")
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage(out, true),
		}, nil
	}

	if cmd.SelectIndex != nil { // TODO: find a more generic approach
		i.log.Info("A specific line was selected", zap.Int("idx", *cmd.SelectIndex))
		cmdTemplate.Message.Select.ItemIdx = *cmd.SelectIndex
		cmdTemplate.Message.Select.Replace = true
	}
	render, err := i.renderer.Get(cmdTemplate.Command.Parser)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}
	message, err := render.RenderMessage(cmd.ToExecute, out, &cmdTemplate)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}
	return executor.ExecuteOutput{
		Message: message,
	}, nil
}

func runCmd(ctx context.Context, tmp osx.TmpDir, in string) (string, error) {
	path, custom := tmp.Get()
	if custom {
		defer os.Setenv("PLUGIN_DEPENDENCY_DIR", os.Getenv("PLUGIN_DEPENDENCY_DIR"))
		os.Setenv("PLUGIN_DEPENDENCY_DIR", path)
	}
	fmt.Println(path)
	fmt.Println(custom)
	fmt.Println(in)

	out, err := pluginx.ExecuteCommand(ctx, in)
	if err != nil {
		return "", err
	}
	return color.ClearCode(out), nil
}