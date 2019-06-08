package main

import (
	"github.com/justmiles/ssm-run-command/cmd"
)

// Version of ssm-run-command. Overwritten during build
var Version = "development"

func main() {
	cmd.Execute(Version)
}
