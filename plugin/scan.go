package plugin

import "strings"

type scan struct {
	Results []result `json:"results"`
}

type result struct {
	Path  string `json:"path"`
	Extra extra  `json:"extra"`
	Start start  `json:"start"`
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

func (s *scan) renderPaths(rootPath string) *scan {
	if len(s.Results) > 0 {
		for _, r := range s.Results {
			r.Path = strings.TrimPrefix(r.Path, rootPath)
		}
	}
	return s
}
