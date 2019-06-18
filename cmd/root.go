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

		fmt.Println(c.Status())
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

	rootCmd.PersistentFlags().IntVar(&c.TargetLimit, "target-limit", 50, `limit execution to first n targets. Max 50`)
	rootCmd.PersistentFlags().IntVar(&c.ExecutionTimeout, "execution-timeout", 3600, `The time in seconds for a command to complete before it is considered to have failed. Default is 3600 (1 hour). Maximum is 172800 (48 hours).`)

	rootCmd.PersistentFlags().StringArrayVar(&c.Targets, "target", nil, `target instances with these values. 
	Example: --target "tag:App=MyApplication" --target "tag:Environment=qa"`)

	// rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Return the command to run against the instances it would invoke against")

}
