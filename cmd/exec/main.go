package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/alexflint/go-arg"
	"github.com/gookit/color"
	"github.com/hashicorp/go-plugin"
	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/executor"
	"github.com/kubeshop/botkube/pkg/pluginx"
	"github.com/mattn/go-shellwords"
	"gopkg.in/yaml.v3"

	x "go.szostok.io/botkube-plugins/internal/exec"
	"go.szostok.io/botkube-plugins/internal/exec/output"
	"go.szostok.io/botkube-plugins/internal/formatx"
	"go.szostok.io/botkube-plugins/internal/getter"
)

// version is set via ldflags by GoReleaser.
var version = "dev"

const (
	pluginName = "x"
)

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
		return install(tool)
	}
	return executor.ExecuteOutput{
		Message: api.NewPlaintextMessage("Command not supported", false),
	}, nil
}

func escapePositionals(in string) string {
	for _, name := range []string{"run", "install"} {
		if strings.Contains(in, name) {
			return strings.Replace(in, name, fmt.Sprintf("%s -- ", name), 1)
		}
	}

	return in
}

func install(tool string) (executor.ExecuteOutput, error) {
	fmt.Println("in tool", tool)
	cmd := fmt.Sprintf("/tmp/bin/eget --to=/tmp/bin %s ", tool)
	fmt.Printf("running %s\n", cmd)
	err := os.MkdirAll("/tmp/bin", 0o777)
	if err != nil {
		log.Println("while creating it", err)
		return executor.ExecuteOutput{}, err
	}

	out, err := runCmdInstall(fmt.Sprintf("wget -O /tmp/bin/eget https://github.com/mszostok/botkube/releases/download/v0.66.0/eget-%s", runtime.GOOS))
	if err != nil {
		log.Println("while downloading it", out, err)
		return executor.ExecuteOutput{}, err
	}

	out, err = runCmdInstall("chmod +x /tmp/bin/eget")
	if err != nil {
		log.Println("while changing mod", out, err)
		return executor.ExecuteOutput{}, err
	}

	out, err = runCmdTool(cmd)
	if err != nil {
		log.Println("while running mod", out, err)
		return executor.ExecuteOutput{}, err
	}

	return executor.ExecuteOutput{
		Message: api.NewPlaintextMessage("Installed successfully", false),
	}, nil
}

func runCmdTool(in string) (string, error) {
	args, err := shellwords.Parse(in)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PATH=$PATH:/tmp/bin")

	out, err := cmd.CombinedOutput()
	return string(out), err
}

var hasher = sha256.New()

func sha(in string) string {
	hasher.Reset()
	hasher.Write([]byte(in))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
func run(ctx context.Context, cfgs []*executor.Config, tool string) (executor.ExecuteOutput, error) {
	var cfg x.Config
	err := pluginx.MergeExecutorConfigs(cfgs, &cfg)
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	cmd := x.Parse(tool)
	out, err := runCmd(cmd.ToExecute)
	if err != nil {
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage(fmt.Sprintf("%s\n%s", out, err.Error()), false),
		}, nil
	}

	if cmd.IsRawRequired {
		return executor.ExecuteOutput{
			Message: api.NewCodeBlockMessage(out, true),
		}, nil
	}

	for _, tpl := range cfg.Interactive.Templates {
		err := getter.Download(ctx, tpl.Ref, filepath.Join("tmp", "x-templates", sha(tpl.Ref)))
		if err != nil {
			return executor.ExecuteOutput{}, err
		}
	}

	var interactiveTemplates x.Interactive
	err = filepath.WalkDir(filepath.Join("tmp", "x-templates"), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		fmt.Println(filepath.Ext(d.Name()))
		if filepath.Ext(d.Name()) != "yaml" {
			return nil
		}

		file, err := os.ReadFile(d.Name())
		if err != nil {
			return err
		}

		var cfg x.Interactive
		err = yaml.Unmarshal(file, &cfg)
		if err != nil {
			return err
		}
		interactiveTemplates.Interactive = append(interactiveTemplates.Interactive, cfg.Interactive...)
		return nil
	})
	if err != nil {
		return executor.ExecuteOutput{}, err
	}

	interactivityConfig, found := interactiveTemplates.FindWithPrefix(cmd.ToExecute)
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

func runCmd(in string) (string, error) {
	args, err := shellwords.Parse(in)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(fmt.Sprintf("/tmp/bin/%s", args[0]), args[1:]...)
	fmt.Println(cmd.String())
	out, err := cmd.CombinedOutput()
	return color.ClearCode(string(out)), err
}

func runCmdInstall(in string) (string, error) {
	args, err := shellwords.Parse(in)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
