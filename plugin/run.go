package plugin

import (
	"log"
	"os"

	"github.com/cocov-ci/go-plugin-kit/cocov"
	"go.uber.org/zap"
)

const (
	noYaml     = "auto"
	cmd        = "semgrep"
	maxWorkers = 4
)

func Run(ctx cocov.Context) error {
	s := newRunner(ccExec{})
	l, err := setupLogger()
	if err != nil {
		log.Fatalf("error configuring logger %s", err.Error())
	}

	res, err := s.run(ctx, l)
	if err != nil {
		return err
	}

	for _, r := range res {
		err = ctx.EmitIssue(
			r.kind, r.Path,
			r.startLine(), r.startLine(),
			r.message(), r.hashID(),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func setupLogger() (*zap.Logger, error) {
	l, err := zap.NewProduction()
	if os.Getenv("COCOV_ENV") == "development" {
		opts := zap.Development()
		l, err = zap.NewDevelopment(opts)
	}
	if err != nil {
		return nil, err
	}
	return l.With(zap.String("plugin", "semgrep")), nil
}
