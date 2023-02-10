package plugin

import (
	"encoding/json"
)

type job struct {
	rootPath string
	path     string
	args     []string
}

func newJob(rootPath, path string, args []string) job {
	return job{rootPath: rootPath, path: path, args: args}
}

type execFn func(args []string, path string) ([]byte, error)

func (j job) run(fn execFn, resChan chan *results, errChan chan error) {
	out, err := fn(j.args, j.path)
	if err != nil {
		errChan <- err
		return
	}

	res := &results{}
	if err = json.Unmarshal(out, res); err != nil {
		errChan <- decodeErr(j.path, j.args, out, err)
	}

	if len(res.Results) > 0 {
		resChan <- res.renderResults(j.rootPath)
	}

}
