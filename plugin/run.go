package plugin

import (
	"github.com/cocov-ci/go-plugin-kit/cocov"
)

const (
	noYaml     = "auto"
	cmd        = "semgrep"
	maxWorkers = 4
)

func Run(ctx cocov.Context) error {
	s := newRunner(ccExec{})
	res, err := s.run(ctx, nil)
	if err != nil {
		return err
	}

	_ = res

	return nil
}
