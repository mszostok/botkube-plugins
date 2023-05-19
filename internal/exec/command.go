package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-shellwords"
)

// ExecuteCommand is a simple wrapper around exec.CommandContext to simplify running a given
// command.
func ExecuteCommand(ctx context.Context, rawCmd string) (string, error) {
	return ExecuteCommandWithEnvs(ctx, rawCmd, nil)
}

// ExecuteCommandWithEnvs is a simple wrapper around exec.CommandContext to simplify running a given
// command.
func ExecuteCommandWithEnvs(ctx context.Context, rawCmd string, envs map[string]string) (string, error) {
	var stdout, stderr bytes.Buffer

	parser := shellwords.NewParser()
	parser.ParseEnv = false
	parser.ParseBacktick = false
	args, err := parser.Parse(rawCmd)
	if err != nil {
		return "", err
	}

	if len(args) < 1 {
		return "", fmt.Errorf("invalid raw command: %q", rawCmd)
	}

	bin, binArgs := args[0], args[1:]
	depDir, found := os.LookupEnv("PLUGIN_DEPENDENCY_DIR")
	if found {
		// Use exactly the binary from the $PLUGIN_DEPENDENCY_DIR directory
		bin = fmt.Sprintf("%s/%s", depDir, bin)
	}

	//nolint:gosec // G204: Subprocess launched with a potential tainted input or cmd arguments
	cmd := exec.CommandContext(ctx, bin, binArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	cmd.Env = append(cmd.Env, os.Environ()...)

	for key, value := range envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if err = cmd.Run(); err != nil {
		return "", runErr(stdout.String(), stderr.String(), err)
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		return "", fmt.Errorf("got non-zero exit code, stdout [%q], stderr [%q]", stdout.String(), stderr.String())
	}
	return stdout.String(), nil
}

func runErr(sout, serr string, err error) error {
	strBldr := strings.Builder{}
	if sout != "" {
		strBldr.WriteString(sout)
		strBldr.WriteString("\n")
	}

	if serr != "" {
		strBldr.WriteString(serr)
		strBldr.WriteString("\n")
	}

	return fmt.Errorf("%s%w", strBldr.String(), err)
}
