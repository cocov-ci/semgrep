package plugin

import (
	"fmt"
	"github.com/cocov-ci/go-plugin-kit/cocov"
	"path/filepath"
	"strings"
)

// Those constants represents official semgrep results categories.
const (
	bestPractice    = "best-practice"
	correctness     = "correctness"
	maintainability = "maintainability"
	performance     = "performance"
	portability     = "portability"
	security        = "security"
)

var cocovIssues = map[string]cocov.IssueKind{
	correctness:     cocov.IssueKindQuality,
	portability:     cocov.IssueKindQuality,
	maintainability: cocov.IssueKindQuality,
	bestPractice:    cocov.IssueKindConvention,
	performance:     cocov.IssueKindPerformance,
	security:        cocov.IssueKindSecurity,
}

type results struct {
	Results []*result `json:"results"`
}

type result struct {
	Path  string `json:"path"`
	Extra extra  `json:"extra"`
	Start start  `json:"start"`
	kind  cocov.IssueKind
	valid bool
}

type extra struct {
	Message  string   `json:"message"`
	Metadata metadata `json:"metadata"`
}

type metadata struct {
	Category   string   `json:"category"`
	References []string `json:"references"`
}

type start struct {
	Line int `json:"line"`
}

func (r *results) renderResults(rootPath string) *results {
	if len(r.Results) > 0 {
		for _, res := range r.Results {
			v, ok := cocovIssues[res.Extra.Metadata.Category]
			res.valid = ok
			if !ok {
				continue
			}

			res.kind = v

			refs := res.Extra.Metadata.References
			if len(refs) > 0 {
				res.Extra.Message = renderMessage(res.Extra.Message, refs)
			}

			if filepath.Dir(res.Path) == rootPath {
				res.Path = filepath.Base(res.Path)
				continue
			}

			res.Path = strings.TrimPrefix(res.Path, rootPath)[1:]

		}
	}
	return r
}

func renderMessage(message string, references []string) string {
	return fmt.Sprintf("%s\nreferences:\n%s", message, strings.Join(references, "\n"))
}
