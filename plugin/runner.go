package plugin

import (
	"encoding/json"
	"sync"

	"github.com/cocov-ci/go-plugin-kit/cocov"
	"go.uber.org/zap"
)

type runner struct{ executor }

func newRunner(e executor) *runner { return &runner{e} }

func (ru *runner) run(ctx cocov.Context) ([]*result, error) {
	rootPath := ctx.Workdir()
	rootYaml := noYaml

	configs, err := findYamlRecursive(rootPath)
	if err != nil {
		ctx.L().Error("Error looking for semgrep configuration files", zap.Error(err))
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
		return ru.parallel(jobs, ctx.L())
	}

	out, err := ru.exec(rootArgs, rootPath)
	if err != nil {
		ctx.L().Error("Error", zap.Error(err))
		return nil, err
	}

	res := &results{}
	if err = json.Unmarshal(out, res); err != nil {
		decodeError := decodeErr(rootPath, rootArgs, out, err)
		ctx.L().Error("Error", zap.Error(decodeError))
		return nil, decodeError
	}

	res.renderResults(rootPath)

	return res.Results, nil
}

func (ru *runner) parallel(jobs []job, logger *zap.Logger) ([]*result, error) {
	errs := make([]error, 0, len(jobs))
	res := make([]*results, 0, len(jobs))

	jobChan := make(chan job)
	resChan := make(chan *opResult)
	doneChan := make(chan bool)

	wg := &sync.WaitGroup{}
	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go func(done func()) {
			for j := range jobChan {
				j.run(ru.exec, resChan)
			}
			done()
		}(wg.Done)
	}

	go func() {
		for r := range resChan {
			if r.error != nil {
				errs = append(errs, r.error)
			} else {
				res = append(res, r.results)
			}
		}
		close(doneChan)
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

	<-doneChan

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

func buildJobs(configs []projectConfig, rootPath string, rootArgs []string) []job {
	jobs := make([]job, 0, len(configs))

	for _, config := range configs {
		if config.path == rootPath {
			continue
		}

		args := newArgs(config.filePath(), config.path)
		j := newJob(rootPath, config.path, args)
		jobs = append(jobs, j)
		rootArgs = append(rootArgs, "--exclude", config.path)
	}

	jobs = append(jobs, newJob(rootPath, rootPath, rootArgs))
	return jobs
}
