package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cocov-ci/go-plugin-kit/cocov"
	"github.com/stretchr/testify/require"
)

func TestFindYamlRecursive(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	wd = findRepositoryRoot(t, wd, "semgrep")

	t.Run("Yaml not found", func(t *testing.T) {
		parent := filepath.Join(wd, "mocks")
		paths, err := findYamlRecursive(parent)
		require.NoError(t, err)
		require.Len(t, paths, 0)
	})

	t.Run("Found yaml", func(t *testing.T) {
		parent := filepath.Join(wd, "plugin", "fixtures")
		pathsFound, err := findYamlRecursive(parent)
		require.NoError(t, err)

		entries, err := os.ReadDir(parent)
		require.NoError(t, err)
		// should avoid test-no-yaml folder
		require.Equal(t, len(entries)-1, len(pathsFound))

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

func findRepositoryRoot(t *testing.T, current, parent string) string {
	out, err := cocov.Exec("git", []string{"rev-parse", "--show-toplevel"}, nil)
	require.NoError(t, err)
	return strings.TrimSpace(string(out))
}
