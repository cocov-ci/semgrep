package plugin

import (
	"errors"
	"fmt"
	"strings"
)

func newArgs(configFile, path string) []string {
	return []string{"scan", "--config", configFile, "--json", path}
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

func formatErrors(errs []error) error {
	finalErr := "\n"
	for i := 0; i < len(errs); i++ {
		finalErr = fmt.Sprintf("%serror %d :\n%s\n", finalErr, i+1, errs[i].Error())
	}

	return errors.New(finalErr)
}
