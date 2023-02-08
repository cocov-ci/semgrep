package plugin

import (
	"io/fs"
	"os"
	"path/filepath"
)

type projectConfig struct {
	file string
	path string
}

func (p projectConfig) filePath() string {
	return filepath.Join(p.path, p.file)
}

func findYamlRecursive(rootPath string) ([]projectConfig, error) {
	root := os.DirFS(rootPath)
	var paths []projectConfig

	err := fs.WalkDir(root, ".",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if d.Name() == "semgrep.yaml" || d.Name() == ".semgrep.yaml" {
				p := projectConfig{
					file: d.Name(),
					path: filepath.Join(rootPath, filepath.Dir(path)),
				}
				paths = append(paths, p)
			}
			return nil
		})

	if err != nil {
		return nil, err
	}

	return paths, nil
}
