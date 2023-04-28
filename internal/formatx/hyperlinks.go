package formatx

import (
	"regexp"
	"strings"

	"github.com/kubeshop/botkube/pkg/formatx"
)

var compiledRegex = regexp.MustCompile(`<(https?://[a-z.0-9/\-_=]*)>`)

// RemoveHyperLinks removes the hyperlink text from url.
func RemoveHyperLinks(in string) string {
	matched := compiledRegex.FindAllStringSubmatch(in, -1)
	if len(matched) >= 1 {
		for _, match := range matched {
			if len(match) == 2 {
				in = strings.ReplaceAll(in, match[0], match[1])
			}
		}
	}
	return formatx.RemoveHyperlinks(in)
}

func Normalize(in string) string {
	out := RemoveHyperLinks(in)
	out = strings.NewReplacer(`“`, `"`, `”`, `"`, `‘`, `"`, `’`, `"`).Replace(out)

	out = strings.TrimSpace(out)

	return out
}
