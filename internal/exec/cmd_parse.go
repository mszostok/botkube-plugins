package exec

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.szostok.io/botkube-plugins/internal/ptr"
)

var compiledRegex = regexp.MustCompile(fmt.Sprintf(`%s:(\d+)`, SelectIndexIndicator))

const (
	NoProcessingIndicator = "@no-interactivity"
	SelectIndexIndicator  = "@idx"
)

type Command struct {
	ToExecute     string
	IsRawRequired bool
	SelectIndex   *int
}

func Parse(cmd string) Command {
	out := Command{
		ToExecute: cmd,
	}
	if strings.Contains(out.ToExecute, NoProcessingIndicator) {
		out.ToExecute = strings.ReplaceAll(out.ToExecute, NoProcessingIndicator, "")
		out.IsRawRequired = true
	}

	out.ToExecute, out.SelectIndex = separateItemIdxAndCommand(out.ToExecute)
	out.ToExecute = strings.TrimSpace(out.ToExecute)

	return out
}

func separateItemIdxAndCommand(cmd string) (cmdToExecute string, idx *int) {
	matched := compiledRegex.FindStringSubmatch(cmd)
	if len(matched) == 2 {
		cmd = strings.Replace(cmd, matched[0], "", 1)
		val, _ := strconv.Atoi(matched[1])
		return cmd, ptr.FromType(val)
	}

	return cmd, nil
}
