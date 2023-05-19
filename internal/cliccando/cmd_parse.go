package cliccando

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var pageIndicatorFinder = regexp.MustCompile(fmt.Sprintf(`%s(\d+)`, PageIndexIndicator))

const (
	PageIndexIndicator = "@page:"
)

type Command struct {
	ToExecute string
	PageIndex int
}

func Parse(pluginName string, cmd string) Command {
	cmd = strings.TrimPrefix(cmd, pluginName)
	out := Command{
		ToExecute: cmd,
	}
	groups := pageIndicatorFinder.FindAllStringSubmatch(cmd, -1)
	if len(groups) > 0 && len(groups[0]) > 1 {
		out.PageIndex, _ = strconv.Atoi(groups[0][1])
	}

	out.ToExecute = pageIndicatorFinder.ReplaceAllString(out.ToExecute, "")
	out.ToExecute = strings.TrimSpace(out.ToExecute)

	return out
}
