package exec

import (
	"context"
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/executor"
	"go.uber.org/zap"

	"go.szostok.io/botkube-plugins/internal/exec/template"
	"go.szostok.io/botkube-plugins/internal/osx"
	"go.szostok.io/botkube-plugins/internal/state"
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

func (i *Runner) Run(ctx context.Context, cfg Config, state *state.Container, tool string) (executor.ExecuteOutput, error) {
	cmd := Parse(tool)

	templates, err := template.Load(ctx, cfg.TmpDir.GetDirectory(), cfg.Templates)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	for _, tpl := range templates.Templates {
		i.log.Info("Command template", zap.String("trigger", tpl.Trigger.Command), zap.String("type", tpl.Type))
	}

	cmdTemplate, found := templates.FindWithPrefix(cmd.ToExecute)

	var out string
	if !found || cmdTemplate.Type != "tutorial" {
		out, err = runCmd(ctx, cfg.TmpDir, cmd.ToExecute)
		if err != nil {
			i.log.Error("failed to run command", zap.String("command", cmd.ToExecute), zap.Error(err))
			return executor.ExecuteOutput{}, err
		}
	}

	if cmd.IsRawRequired {
		i.log.Info("Raw output was explicitly requested")
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage(out, true),
		}, nil
	}

	if !found {
		i.log.Info("Templates config not found for command")
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage(color.ClearCode(out), true),
		}, nil
	}

	render, err := i.renderer.Get(cmdTemplate.Type) // Message.Type
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	cmdTemplate.TutorialMessage.Paginate.CurrentPage = cmd.PageIndex
	message, err := render.RenderMessage(cmd.ToExecute, out, state, &cmdTemplate)
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

	//out, err := pluginx.ExecuteCommand(ctx, in)
	out, err := ExecuteCommand(ctx, in)
	if err != nil {
		return "", err
	}
	return color.ClearCode(out), nil
}
