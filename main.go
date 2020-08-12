package main

import (
	"github.com/justmiles/ssm-run-command/cmd"
)

// Version of ssm-run-command. Overwritten during build
var version = "development"

func main() {
	cmd.Execute(version)
}
