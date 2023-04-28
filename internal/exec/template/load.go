package template

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(ctx context.Context, tmpDir string, templateSources []Source) (Config, error) {
	dir := filepath.Join(tmpDir, "x-templates")
	err := EnsureDownloaded(ctx, templateSources, dir)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

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
