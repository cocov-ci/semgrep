package plugin

import "encoding/json"

type job struct {
	rootPath string
	path     string
	args     []string
}

func newJob(rootPath, path string, args []string) job {
	return job{rootPath: rootPath, path: path, args: args}
}

type execFn func(args []string, path string) ([]byte, error)

func (j job) run(fn execFn, resChan chan *opResult) {
	out, err := fn(j.args, j.path)
	if err != nil {
		resChan <- &opResult{error: err}
		return
	}

	res := &results{}
	if err = json.Unmarshal(out, res); err != nil {
		resChan <- &opResult{error: decodeErr(j.path, j.args, out, err)}
	}

	if len(res.Results) > 0 {
		resChan <- &opResult{results: res.renderResults(j.rootPath)}
	}

}
