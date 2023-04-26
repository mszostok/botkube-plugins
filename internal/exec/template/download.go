package template

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"path/filepath"

	"go.szostok.io/botkube-plugins/internal/exec"
	"go.szostok.io/botkube-plugins/internal/getter"
	"go.szostok.io/botkube-plugins/internal/osx"
)

var hasher = sha256.New()

func sha(in string) string {
	hasher.Reset()
	hasher.Write([]byte(in))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
func EnsureDownloaded(ctx context.Context, templateSources []exec.Template, dir string) error {
	for _, tpl := range templateSources {
		dst := filepath.Join(dir, sha(tpl.Ref))
		err := osx.RunIfFileDoesNotExist(dst, func() error {
			return getter.Download(ctx, tpl.Ref, dst)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
