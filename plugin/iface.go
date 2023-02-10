package plugin

import "github.com/cocov-ci/go-plugin-kit/cocov"

type executor interface {
	Exec2(string, []string, *cocov.ExecOpts) (stdout, stderr []byte, err error)
}

type ccExec struct{}

func (ccExec) Exec2(cmd string, args []string, opts *cocov.ExecOpts) (stdout, stderr []byte, err error) {
	return cocov.Exec2(cmd, args, opts)
}
