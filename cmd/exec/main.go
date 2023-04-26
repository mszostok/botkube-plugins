package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/alexflint/go-arg"
	"github.com/gookit/color"
	"github.com/hashicorp/go-plugin"
	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/executor"
	"github.com/kubeshop/botkube/pkg/pluginx"

	x "go.szostok.io/botkube-plugins/internal/exec"
	"go.szostok.io/botkube-plugins/internal/exec/output"
	"go.szostok.io/botkube-plugins/internal/exec/template"
	"go.szostok.io/botkube-plugins/internal/formatx"
)

// version is set via ldflags by GoReleaser.
var version = "dev"

const (
	pluginName = "x"
	binaryName = "eget"
)

var egetBinaryDownloadLinks = map[string]string{
	"windows/amd64": "https://github.com/zyedidia/eget/releases/download/v1.3.3/eget-1.3.3-windows_amd64.zip//eget-1.3.3-windows_amd64",
	"darwin/amd64":  "https://github.com/zyedidia/eget/releases/download/v1.3.3/eget-1.3.3-darwin_amd64.tar.gz//eget-1.3.3-darwin_amd64",
	"darwin/arm64":  "https://github.com/zyedidia/eget/releases/download/v1.3.3/eget-1.3.3-darwin_arm64.tar.gz//eget-1.3.3-darwin_arm64",
	"linux/amd64":   "https://github.com/zyedidia/eget/releases/download/v1.3.3/eget-1.3.3-linux_amd64.tar.gz//eget-1.3.3-linux_amd64",
	"linux/arm64":   "https://github.com/zyedidia/eget/releases/download/v1.3.3/eget-1.3.3-linux_arm64.tar.gz//eget-1.3.3-linux_arm64",
	"linux/386":     "https://github.com/zyedidia/eget/releases/download/v1.3.3/eget-1.3.3-linux_386.tar.gz//eget-1.3.3-linux_386",
}

// InstallExecutor implements Botkube executor plugin.
type InstallExecutor struct{}

func (i *InstallExecutor) Help(_ context.Context) (api.Message, error) {
	help := heredoc.Doc(`
		Usage:
		  x run [COMMAND] [FLAGS]    Run a specified command with optional flags
		  x install [SOURCE]         Install a binary using the https://github.com/zyedidia/eget syntax.
		
		Usage Examples:
		  x install https://get.helm.sh/helm-v3.10.3-linux-amd64.tar.gz --file helm    # Install the Helm binary
		  x run helm list -A   # Run the 'helm list -A' command. 
		
		Options:
		  -h, --help                 Show this help message`)
	return api.NewCodeBlockMessage(help, true), nil
}

// Metadata returns details about Echo plugin.
func (*InstallExecutor) Metadata(context.Context) (api.MetadataOutput, error) {
	return api.MetadataOutput{
		Version:     "v1.0.0",
		Description: "Runs installed binaries",
		Dependencies: map[string]api.Dependency{
			binaryName: {
				URLs: egetBinaryDownloadLinks,
			},
		},
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

// Execute returns a given command as response.
func (i *InstallExecutor) Execute(ctx context.Context, in executor.ExecuteInput) (executor.ExecuteOutput, error) {
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

	switch {
	case cmd.Run != nil:
		tool := formatx.Normalize(strings.Join(cmd.Run.Tool, " "))
		return run(ctx, in.Configs, tool)
	case cmd.Install != nil:
		tool := formatx.Normalize(strings.Join(cmd.Install.Tool, " "))
		fmt.Println("in tool", tool)

		dir, _ := getInstallDirectory()
		cmd := fmt.Sprintf("eget --to=%s %s ", dir, tool)
		_, err := pluginx.ExecuteCommand(ctx, cmd)
		if err != nil {
			return executor.ExecuteOutput{}, err
		}

		return executor.ExecuteOutput{
			Message: api.NewPlaintextMessage("Binary was installed successfully", false),
		}, nil
	}
	return executor.ExecuteOutput{
		Message: api.NewPlaintextMessage("Command not supported", false),
	}, nil
}

func getInstallDirectory() (string, bool) {
	depDir := os.Getenv("PLUGIN_DEPENDENCY_DIR")
	if depDir != "" {
		return depDir, false
	}

	return "/tmp/bin", true
}

func escapePositionals(in string) string {
	for _, name := range []string{"run", "install"} {
		if strings.Contains(in, name) {
			return strings.Replace(in, name, fmt.Sprintf("%s -- ", name), 1)
		}
	}

	return in
}

func run(ctx context.Context, cfgs []*executor.Config, tool string) (executor.ExecuteOutput, error) {
	var cfg x.Config
	err := pluginx.MergeExecutorConfigs(cfgs, &cfg)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	cmd := x.Parse(tool)
	out, err := runCmd(ctx, cmd.ToExecute)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	if cmd.IsRawRequired {
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage(out, true),
		}, nil
	}

	interactiveTpls, err := template.Load(ctx, cfg.Interactive.Templates)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	interactivityConfig, found := interactiveTpls.FindWithPrefix(cmd.ToExecute)
	if !found {
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage(out, true),
		}, nil
	}

	if cmd.SelectIndex != nil {
		interactivityConfig.Message.Select.ItemIdx = *cmd.SelectIndex
		interactivityConfig.Message.Select.Replace = true
	}

	interactiveMsg, err := output.BuildMessage(cmd.ToExecute, out, interactivityConfig)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}
	return executor.ExecuteOutput{
		Message: interactiveMsg,
	}, nil
}

func main() {
	executor.Serve(map[string]plugin.Plugin{
		pluginName: &executor.Plugin{
			Executor: &InstallExecutor{},
		},
	})
}

func runCmd(ctx context.Context, in string) (string, error) {
	dir, custom := getInstallDirectory()
	if custom {
		in = fmt.Sprintf("%s/%s", dir, in)
	}
	out, err := pluginx.ExecuteCommand(ctx, in)
	if err != nil {
		return "", err
	}
	return color.ClearCode(out), nil
}
