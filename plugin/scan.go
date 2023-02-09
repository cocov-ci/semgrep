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

	bestPractice: cocov.IssueKindConvention,

	performance: cocov.IssueKindPerformance,

	security: cocov.IssueKindSecurity,
}

type scan struct {
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

func (s *scan) renderResults(rootPath string) *scan {
	if len(s.Results) > 0 {
		for _, r := range s.Results {
			v, ok := cocovIssues[r.Extra.Metadata.Category]
			r.valid = ok
			if !ok {
				continue
			}

			r.kind = v

			refs := r.Extra.Metadata.References
			if len(refs) > 0 {
				r.Extra.Message = renderMessage(r.Extra.Message, refs)
			}

			if filepath.Dir(r.Path) != rootPath {
				r.Path = strings.TrimPrefix(r.Path, rootPath)[1:]
			}
		}
	}
	return s
}

func renderMessage(message string, references []string) string {
	return fmt.Sprintf("%s\nreferences:\n%s", message, strings.Join(references, "\n"))
}
