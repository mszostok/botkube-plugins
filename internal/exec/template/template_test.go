package template_test

import (
	"context"
	"testing"

	"github.com/sanity-io/litter"

	"go.szostok.io/botkube-plugins/internal/exec/template"
)

func Test(t *testing.T) {

	dir := t.TempDir()
	load, err := template.Load(context.Background(), dir, []template.Source{
		{
			Ref: "../../../x-templates",
		},
	})
	if err != nil {
		return
	}
	litter.Dump(load.Templates)
	tt, found := load.FindWithPrefix("helm list -A")
	litter.Dump(found)
	litter.Dump(tt)
}
