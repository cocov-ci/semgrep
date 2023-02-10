package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindYamlRecursive(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	wd = findParentDir(t, wd, "semgrep")

	t.Run("Yaml not found", func(t *testing.T) {
		parent := filepath.Join(wd, "mocks")
		paths, err := findYamlRecursive(parent)
		require.NoError(t, err)
		require.Len(t, paths, 0)
	})

	t.Run("Found yaml", func(t *testing.T) {
		parent := filepath.Join(wd, "plugin/fixtures")
		pathsFound, err := findYamlRecursive(parent)
		require.NoError(t, err)

		entries, err := os.ReadDir(parent)
		require.NoError(t, err)
		require.Equal(t, len(entries), len(pathsFound))

		for _, p := range pathsFound {
			info, err := os.Stat(p.path)
			require.NoError(t, err)
			require.True(t, info.IsDir())

			info, err = os.Stat(p.filePath())
			require.NoError(t, err)
			require.False(t, info.IsDir())
		}
	})
}

func findParentDir(t *testing.T, current, parent string) string {
	require.True(t, strings.Contains(current, parent))

	if filepath.Base(current) == parent {
		return current
	}

	return findParentDir(t, filepath.Dir(current), parent)
}
