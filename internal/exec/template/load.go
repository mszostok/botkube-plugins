package template

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"go.szostok.io/botkube-plugins/internal/exec"
)

func Load(ctx context.Context, templateSources []exec.Template) (Config, error) {
	dir := filepath.Join("/tmp", "x-templates")
	err := EnsureDownloaded(ctx, templateSources, dir)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		fmt.Println(filepath.Ext(d.Name()))
		if filepath.Ext(d.Name()) != ".yaml" {
			return nil
		}

		file, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var cfg Config
		err = yaml.Unmarshal(file, &cfg)
		if err != nil {
			return err
		}
		config.Interactive = append(config.Interactive, cfg.Interactive...)
		return nil
	})
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
