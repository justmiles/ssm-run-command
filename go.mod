module github.com/justmiles/ssm-run-command

go 1.12

replace github.com/justmiles/ssm-run-command/cmd => ./cmd

replace github.com/justmiles/ssm-run-command/lib => ./lib

require (
	github.com/aws/aws-sdk-go v1.34.2
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kvz/logstreamer v0.0.0-20150507115422-a635b98146f0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	golang.org/x/text v0.3.1-0.20180807135948-17ff2d5776d2 // indirect
)
