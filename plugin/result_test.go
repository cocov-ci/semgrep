package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderResults(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	expectedPaths := []string{
		"project_root/file.yaml",
		"project_root/project1/file.yaml",
		"project_root/project2/file.go",
		"project_root/project3/file.rb",
		"project_root/project4/unknown.donno",
	}

	categories := []string{
		correctness,
		bestPractice,
		performance,
		security,
		"some-category",
	}

	res := &results{}
	baseMsg := "some-message-"
	baseRef := "some-ref-"

	for i, ep := range expectedPaths {
		p := filepath.Join(wd, ep)
		message := fmt.Sprintf("%s%d", baseMsg, i)
		cat := categories[i]
		refs := []string{
			fmt.Sprintf("%s%d", baseRef, i),
			fmt.Sprintf("%s%d", baseRef, i+i),
		}

		res.Results = append(res.Results, newResult(p, message, cat, refs))
	}

	res.renderResults(wd)

	for i, r := range res.Results {
		v, ok := cocovIssues[r.Extra.Metadata.Category]
		assert.Equal(t, r.valid, ok)

		if r.valid {
			assert.Equal(t, r.Path, expectedPaths[i])
			assert.Equal(t, v, r.kind)

			msg := fmt.Sprintf("%s%d", baseMsg, i)
			refs := r.Extra.Metadata.References
			expectedMsg := renderMessage(msg, refs)
			assert.Equal(t, r.Extra.Message, expectedMsg)
		}
	}
}

func newResult(filePath, message, category string, refs []string) *result {
	return &result{
		Path: filePath,
		Extra: extra{
			Message: message,
			Metadata: metadata{
				Category:   category,
				References: refs},
		},
	}
}
