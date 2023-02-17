package main

import (
	"github.com/cocov-ci/go-plugin-kit/cocov"
	"github.com/cocov-ci/semgrep/plugin"
)

func main() { cocov.Run(plugin.Run) }
