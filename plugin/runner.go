package plugin

import (
	"encoding/json"
	"github.com/cocov-ci/go-plugin-kit/cocov"
	"go.uber.org/zap"
	"sync"
)

type runner struct{ executor }

func newRunner(e executor) *runner { return &runner{e} }

func (ru *runner) run(ctx cocov.Context, logger *zap.Logger) ([]*result, error) {
	rootPath := ctx.Workdir()
	rootYaml := noYaml

	configs, err := findYamlRecursive(rootPath)
	if err != nil {
		logger.Error("Error looking for semgrep configuration files", zap.Error(err))
		return nil, err
	}

	for _, config := range configs {
		if config.path == rootPath {
			rootYaml = config.filePath()
			break
		}
	}

	rootArgs := newArgs(rootYaml, rootPath)

	if len(configs) > 1 {
		jobs := buildJobs(configs, rootPath, rootArgs)
		return ru.parallel(jobs, logger)
	}

	out, err := ru.exec(rootArgs, rootPath)
	if err != nil {
		logger.Error("Error", zap.Error(err))
		return nil, err
	}

	res := &results{}
	if err = json.Unmarshal(out, res); err != nil {
		decodeError := decodeErr(rootPath, rootArgs, out, err)
		logger.Error("Error", zap.Error(decodeError))
		return nil, decodeError
	}

	res.renderResults(rootPath)

	return res.Results, nil
}

func (ru *runner) parallel(jobs []job, logger *zap.Logger) ([]*result, error) {
	errs := make([]error, 0, len(jobs))
	res := make([]*results, 0, len(jobs))
	numWorkers := maxWorkers

	if len(jobs) < maxWorkers {
		numWorkers = len(jobs)
	}

	jobChan := make(chan job)
	errChan := make(chan error)
	resChan := make(chan *results)

	wg := &sync.WaitGroup{}
	wg.Add(numWorkers)

	for i := 0; i < maxWorkers; i++ {
		go func(id int, done func()) {
			for j := range jobChan {
				j.run(ru.exec, resChan, errChan)
			}
			done()
		}(i, wg.Done)
	}

	go func() {
		for r := range resChan {
			res = append(res, r)
		}
	}()

	go func() {
		for err := range errChan {
			if err != nil {
				errs = append(errs, err)
			}
		}
	}()

	infoFields := make([]zap.Field, 0, len(jobs))
	for _, j := range jobs {
		jobChan <- j
		infoFields = append(infoFields, zap.String("path", j.path))
	}

	logger.Info("Current jobs:", infoFields...)

	close(jobChan)

	wg.Wait()
	close(resChan)
	close(errChan)

	if len(errs) > 0 {
		return nil, formatErrors(errs)
	}

	var total []*result
	for _, r := range res {
		total = append(total, r.Results...)
	}

	return total, nil
}

func (ru *runner) exec(args []string, cwd string) ([]byte, error) {
	opts := &cocov.ExecOpts{Workdir: cwd}
	stdOut, stdErr, err := ru.Exec2(cmd, args, opts)
	if err != nil {
		return nil, runErr(cwd, args, stdOut, stdErr, err)
	}
	return stdOut, nil
}
