package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindYamlRecursive(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	t.Run("Yaml not found", func(t *testing.T) {
		paths, err := findYamlRecursive(wd)
		require.NoError(t, err)
		require.Len(t, paths, 0)
	})

	t.Run("Found yaml", func(t *testing.T) {
		paths := []string{"fixture1", "fixture2", "fixture3"}

		for i, p := range paths {
			dir, err := os.MkdirTemp(wd, p)
			require.NoError(t, err)
			paths[i] = dir

			fixturePath := filepath.Join(dir, "semgrep.yaml")
			_, err = os.Create(fixturePath)
			require.NoError(t, err)
		}

		t.Cleanup(func() {
			for _, p := range paths {
				_ = os.RemoveAll(p)
			}
		})

		pathsFound, err := findYamlRecursive(wd)
		require.NoError(t, err)
		assert.Len(t, pathsFound, 3)

		okCount := 0

		for _, p := range paths {
			for _, found := range pathsFound {
				if found.path == p {
					okCount += 1
					f, err := os.Stat(found.filePath())
					assert.NoError(t, err)
					assert.False(t, f.IsDir())
				}
			}
		}

		assert.Equal(t, len(paths), okCount)
	})
}
