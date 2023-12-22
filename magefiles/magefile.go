package main

import (
	"github.com/magefile/mage/mg"
	"go.szostok.io/magex/deps"
	"go.szostok.io/magex/shx"

	"go.szostok.io/tools/target"
)

var (
	Default = Build.PluginsSingle

	Aliases = map[string]interface{}{
		"l": Lint,
	}
)

const (
	GolangciLintVersion = "1.55.2"
	bin                 = "bin"
	pluginDistDirName   = "./plugin-dist"
)

// "Go" Targets

type Build mg.Namespace

// Plugins Builds all plugins for all defined platforms.
func (Build) Plugins() {
	target.GenerateGoreleaserFile()
	shx.MustCmdf(`go run github.com/kubeshop/botkube/hack/target/build-plugins`).MustRunV()
	shx.MustCmdf(`go run github.com/kubeshop/botkube/hack -use-archive=true`).MustRunV()
}

// PluginsSingle Builds all plugins only for current GOOS and GOARCH.
func (Build) PluginsSingle() {
	target.GenerateGoreleaserFile()
	mg.Deps(shx.MustCmdf(`go run github.com/kubeshop/botkube/hack/target/build-plugins --single-platform --output-mode binary`).RunV)
	shx.MustCmdf(`go run github.com/kubeshop/botkube/hack -use-archive=false`).MustRunV()
}

// Lint Runs linters on the codebase
func Lint() error {
	mg.Deps(mg.F(deps.EnsureGolangciLint, bin, GolangciLintVersion))
	return shx.MustCmdf(`./%s/golangci-lint run --fix ./...`, bin).RunV()
}

// "Docs" Targets

type Docs mg.Namespace

// Fmt Formats markdown documentation
func (d Docs) Fmt() error {
	mg.Deps(mg.F(deps.EnsurePrettier, bin))
	return target.FmtDocs(false)
}

// Check Checks formatting and links in *.md files
func (d Docs) Check() error {
	mg.Deps(mg.F(deps.EnsurePrettier, bin))

	return target.FmtDocs(true)
}

// "Test" Targets

type Test mg.Namespace

// Unit Executes Go unit tests.
func (Test) Unit() error {
	return shx.MustCmdf(`go test -coverprofile=coverage.out ./...`).Run()
}

// Coverage Generates file with unit test coverage data and open it in browser
func (t Test) Coverage() error {
	mg.Deps(t.Unit)
	return shx.MustCmdf(`go tool cover -html=coverage.out`).Run()
}
