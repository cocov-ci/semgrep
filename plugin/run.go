package plugin

import (
	"encoding/json"

	"github.com/cocov-ci/go-plugin-kit/cocov"
	"go.uber.org/zap"
)

const (
	noYaml     = "auto"
	cmd        = "semgrep"
	maxWorkers = 4
)

func Run(ctx cocov.Context) error {
	return nil
}

func run(ctx cocov.Context, logger *zap.Logger) ([]*scan, error) {
	rootPath := ctx.Workdir()
	rootYaml := noYaml

	individualConfigs, err := findYamlRecursive(rootPath)
	if err != nil {
		logger.Error("Error looking for semgrep configuration files", zap.Error(err))
		return nil, err
	}

	for _, config := range individualConfigs {
		if config.path == rootPath {
			rootYaml = config.path
			break
		}
	}

	rootArgs := newArgs(rootYaml, rootPath)

	if len(individualConfigs) > 0 {
		jobs := make([]job, 0, len(individualConfigs)+1)

		for _, config := range individualConfigs {
			rootArgs = append(rootArgs, "--exclude", config.path)
			args := newArgs(config.filePath(), config.path)
			j := newJob(rootPath, config.path, args)
			jobs = append(jobs, j)
		}

		jobs = append(jobs, newJob(rootPath, rootPath, rootArgs))
		return runParallel(jobs, logger)
	}

	out, err := runScan(rootArgs, rootPath)
	if err != nil {
		logger.Error("Error", zap.Error(err))
		return nil, err
	}

	s := &scan{}
	if err = json.Unmarshal(out, s); err != nil {
		decodeError := decodeErr(rootPath, rootArgs, out, err)
		logger.Error("Error", zap.Error(decodeError))
		return nil, decodeError
	}

	s.renderPaths(rootPath)

	return []*scan{s}, nil
}
