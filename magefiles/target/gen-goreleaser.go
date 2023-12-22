package target

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"text/template"

	"github.com/samber/lo"
	"go.szostok.io/magex/printer"
)

const (
	templateFile = "./.goreleaser.plugin.tpl.yaml"
	outputFile   = "./.goreleaser.plugin.yaml"
	entrypoint   = "./cmd"
	filePerm     = 0o644

	fileNoEditHeader = "# The code has been automatically generated and should not be modified directly. To update, run 'mage build:plugins' from the root directory of this repository."
)

type (
	Plugins []Plugin
	Plugin  struct {
		Name string
		Type string
	}
)

func GenerateGoreleaserFile() {
	printer.Cmd("Re-generating .goreleaser.plugin.yaml file")
	var (
		executors = lo.Must(ignoreNotExistError(os.ReadDir(entrypoint + "/executor")))
		sources   = lo.Must(ignoreNotExistError(os.ReadDir(entrypoint + "/source")))
	)

	var plugins Plugins
	for _, d := range executors {
		plugins = append(plugins, Plugin{
			Type: "executor",
			Name: d.Name(),
		})
	}
	for _, d := range sources {
		plugins = append(plugins, Plugin{
			Type: "source",
			Name: d.Name(),
		})
	}

	file := lo.Must(os.ReadFile(templateFile))

	//  Change delims to not interfere with the GoReleaser templates.
	tpl := lo.Must(template.New("goreleaser").Delims("<", ">").Parse(string(file)))

	dst := lo.Must(os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePerm))

	fmt.Fprintln(dst, fileNoEditHeader)
	fmt.Fprintln(dst)

	lo.Must0(tpl.Execute(dst, plugins))
}

func ignoreNotExistError[T any](val T, err error) (T, error) {
	if errors.Is(err, fs.ErrNotExist) {
		return val, nil
	}
	return val, err
}
