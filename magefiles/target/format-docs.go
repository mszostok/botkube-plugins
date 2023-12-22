package target

import (
	"go.szostok.io/magex/shx"
)

func FmtDocs(onlyCheck bool) error {
	return shx.MustCmdf(`./bin/node_modules/.bin/prettier --write  "**/*.md" %s`,
		WithOptArg("--check", onlyCheck),
	).RunV()
}

func WithOptArg(key string, shouldAdd bool) string {
	if shouldAdd {
		return key
	}
	return ""
}
