package exec

import (
	"fmt"
	"regexp"
	"strings"
)

var selectIndicatorFinder = regexp.MustCompile(fmt.Sprintf(`%s(\d+)`, SelectIndexIndicator))

const (
	RawOutputIndicator   = "@raw"
	SelectIndexIndicator = "@idx:"
)

type Command struct {
	ToExecute     string
	IsRawRequired bool
}

func Parse(cmd string) Command {
	out := Command{
		ToExecute: cmd,
	}
	if strings.Contains(out.ToExecute, RawOutputIndicator) {
		out.ToExecute = strings.ReplaceAll(out.ToExecute, RawOutputIndicator, "")
		out.IsRawRequired = true
	}

	out.ToExecute = selectIndicatorFinder.ReplaceAllString(out.ToExecute, "")
	out.ToExecute = strings.TrimSpace(out.ToExecute)

	return out
}
