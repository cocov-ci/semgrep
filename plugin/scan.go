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
			r.Path = strings.TrimPrefix(r.Path, rootPath)
		}
	}
	return s
}
