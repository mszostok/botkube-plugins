package install

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/executor"
	"github.com/mattn/go-shellwords"
)

func Install(tool string) (executor.ExecuteOutput, error) {

	fmt.Println("in tool", tool)
	cmd := fmt.Sprintf("/tmp/bin/eget --to=/tmp/bin %s ", tool)
	out, err := runCmdTool(cmd)
	if err != nil {
		log.Println("while running mod", out, err)
		return executor.ExecuteOutput{}, err
	}

	return executor.ExecuteOutput{
		Message: api.NewPlaintextMessage("Installed successfully", false),
	}, nil
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
