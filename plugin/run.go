package plugin

import "github.com/cocov-ci/go-plugin-kit/cocov"

const (
	noYaml     = "auto"
	cmd        = "semgrep"
	maxWorkers = 4
)

func Run(ctx cocov.Context) error {
	s := newRunner(ccExec{})
	res, err := s.run(ctx)
	if err != nil {
		return err
	}

	for _, r := range res {
		err = ctx.EmitIssue(
			r.kind, r.Path,
			r.startLine(), r.startLine(),
			r.message(), r.hashID(ctx.CommitSHA()),
		)

		if err != nil {
			return err
		}
	}

	return nil
}
