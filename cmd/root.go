package cmd

import (
	"fmt"
	"log"
	"os"

	command "github.com/justmiles/ssm-run-command/lib"
	"github.com/spf13/cobra"
)

var (
	c command.Command
)

var rootCmd = &cobra.Command{
	Use:   "run-command",
	Short: "run a remote command",
	Long:  "Invoke a remote command(s) using SSM RunCommand and stream results back to stderr/stdout",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			log.Fatal("Pass something to run")
		}

		c.Command = args
		exitCode, err := c.Run()
		if err != nil {
			fmt.Println(err)
		}

		os.Exit(exitCode)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}
`)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.Flags().SetInterspersed(false)

	rootCmd.PersistentFlags().IntVarP(&c.TargetLimit, "target-limit", "t", 50, "(Optional) Limit execution to first n targets")
	rootCmd.PersistentFlags().IntVar(&c.ExecutionTimeout, "execution-timeout", 3600, "(Optional) The time in seconds for a command \nto complete before it is considered to\nhave failed. Default is 3600 (1 hour). Maximum is 172800 (48 hours).")
	rootCmd.PersistentFlags().StringVar(&c.MaxConcurrency, "max-concurrency", "50", "(Optional) The maximum number of instances that \nare allowed to run the command at the same time. You can \nspecify a number such as 10 or a percentage such as 10%.")
	rootCmd.PersistentFlags().StringVar(&c.MaxErrors, "max-errors", "1", "(Optional) The maximum number of errors allowed without \nthe command failing. When the command fails one more time beyond the value \nof MaxErrors, the systems stopnsending the command to additional targets. \nYou can specify a number like 10 or a percentage like 10%.")
	rootCmd.PersistentFlags().StringVarP(&c.Comment, "comment", "c", "invoked using ssm-run-command CLI", "(Optional) Comment for command visible \non the SSM dashboard")
	rootCmd.PersistentFlags().StringVarP(&c.LogGroup, "log-group", "l", "/ssm-run-command", "(Optional) The AWS CloudWatch log group \nfor RunCommand to log to")
	rootCmd.PersistentFlags().StringArrayVar(&c.Targets, "target", nil, "Target instances with these values. \nFor example: --target tag:App=MyApplication --target tag:Environment=qa")

}
