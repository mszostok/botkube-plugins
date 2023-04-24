package exec

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.szostok.io/botkube-plugins/internal/ptr"
)

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

	out.ToExecute, out.SelectIndex = getItemIdx(out.ToExecute)
	out.ToExecute = strings.TrimSpace(out.ToExecute)

	return out
}

var compiledRegex = regexp.MustCompile(fmt.Sprintf(`%s:(\d+)`, SelectIndexIndicator))

func getItemIdx(in string) (string, *int) {
	matched := compiledRegex.FindStringSubmatch(in)
	if len(matched) == 2 {
		in = strings.Replace(in, matched[0], "", 1)
		val, _ := strconv.Atoi(matched[1])
		return in, ptr.FromType(val)
	}

	return in, nil
}
