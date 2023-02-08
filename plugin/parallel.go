package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type job struct {
	rootPath string
	path     string
	args     []string
}

func newJob(rootPath, path string, args []string) job {
	return job{rootPath: rootPath, path: path, args: args}
}

func (j job) run(scanChan chan *scan, errChan chan error) {
	out, err := runScan(j.args, j.path)
	if err != nil {
		errChan <- err
		return
	}

	s := &scan{}
	if err = json.Unmarshal(out, s); err != nil {
		errChan <- decodeErr(j.path, j.args, out, err)
	}

	if len(s.Results) > 0 {
		scanChan <- s.renderPaths(j.rootPath)
	}

}

func runParallel(jobs []job, logger *zap.Logger) ([]*scan, error) {
	errs := make([]error, 0, len(jobs))
	scans := make([]*scan, 0, len(jobs))
	numWorkers := maxWorkers

	if len(jobs) < maxWorkers {
		numWorkers = len(jobs)
	}

	jobChan := make(chan job)
	errChan := make(chan error)
	scanChan := make(chan *scan)

	wg := &sync.WaitGroup{}
	wg.Add(numWorkers)

	for i := 0; i < maxWorkers; i++ {
		go func(id int, done func()) {
			for j := range jobChan {
				s := time.Now()
				logger.Info("Starting scan", zap.String("at path", j.path))
				j.run(scanChan, errChan)
				logger.Info("Scan took", zap.String("total: ", time.Since(s).String()))
			}
			done()
		}(i, wg.Done)
	}

	go func() {
		for s := range scanChan {
			scans = append(scans, s)
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
	close(scanChan)
	close(errChan)

	if len(errs) > 0 {
		return nil, formatErrors(errs)
	}

	return scans, nil
}

func formatErrors(errs []error) error {
	finalErr := "\n"
	for i := 0; i < len(errs); i++ {
		finalErr = fmt.Sprintf("%serror %d :\n%s\n", finalErr, i+1, errs[i].Error())
	}

	return errors.New(finalErr)
}
