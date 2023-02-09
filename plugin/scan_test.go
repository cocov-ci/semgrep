package plugin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func TestRenderResults(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	fmt.Println(wd)

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

	s := &scan{}
	for i, ep := range expectedPaths {
		s.Results = append(s.Results, &result{
			Path: path.Join(wd, ep),
			Extra: extra{
				Message: "some message",
				Metadata: metadata{
					Category: categories[i],
					References: []string{
						"ref 1",
						"ref 2",
					}}},
		})
	}

	s.renderResults(wd)

	expectedMessage := "some message\nreferences:\nref 1\nref 2"

	for i, r := range s.Results {
		v, ok := cocovIssues[r.Extra.Metadata.Category]
		assert.Equal(t, r.valid, ok)

		if r.valid {
			assert.Equal(t, r.Path, expectedPaths[i])
			assert.Equal(t, r.Extra.Message, expectedMessage)
			assert.Equal(t, v, r.kind)
		}
	}
}
