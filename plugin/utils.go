package plugin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cocov-ci/go-plugin-kit/cocov"
)

func newArgs(configFile, path string) []string {
	return []string{"scan", "--config", configFile, "--json", path}
}

func runScan(args []string, cwd string) ([]byte, error) {
	opts := &cocov.ExecOpts{Workdir: cwd}
	stdOut, stdErr, err := cocov.Exec2(cmd, args, opts)
	if err != nil {
		return nil, runErr(cwd, args, stdOut, stdErr, err)
	}
	return stdOut, nil
}

func runErr(path string, args []string, stdOut, stdErr []byte, err error) error {
	errMsg := []string{
		fmt.Sprintf("error running %s %s", cmd, strings.Join(args, " ")),
		fmt.Sprintf("project path: %s", path),
		fmt.Sprintf("error: %s", err.Error()),
		fmt.Sprintf("stdErr: :%s", string(stdErr)),
		fmt.Sprintf("stdOut: :%s", string(stdOut)),
	}

	return errors.New(strings.Join(errMsg, "\n"))
}

func decodeErr(projectPath string, args []string, stdOut []byte, err error) error {
	errMsg := []string{
		"decoding output fails",
		fmt.Sprintf("project path: %s", projectPath),
		fmt.Sprintf("command: semgrep %s", strings.Join(args, " ")),
		fmt.Sprintf("output: :%s", string(stdOut)),
		fmt.Sprintf("error: %s", err.Error()),
	}

	return errors.New(strings.Join(errMsg, "\n"))
}
