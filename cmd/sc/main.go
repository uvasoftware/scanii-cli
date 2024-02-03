package main

import (
	"fmt"
	"github.com/google/gops/agent"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"scanii-cli/cmd/sc/commands"
)

var verbose bool

func init() {

}

func main() {

	rootCmd := &cobra.Command{
		Use:     "sc",
		Version: "0.0.1",
		Short:   "Scanii CLI",
		Long:    "A CLI to help you integrate Scanii (https://www.scanii.com) with your application",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}

			handler := tint.NewHandler(os.Stdout, &tint.Options{
				AddSource:  true,
				Level:      level,
				TimeFormat: "2006/01/02 15:04",
				NoColor:    false,
			})
			slog.SetDefault(slog.New(handler))
			slog.Debug("running in debug mode")

			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := agent.Listen(agent.Options{
				ShutdownCleanup: true,
			}); err != nil {
				panic(err)
			}

		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print application version and runtime information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello version")
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(commands.ConfigureCommand())
	rootCmd.AddCommand(commands.PingCommand())
	rootCmd.AddCommand(commands.ServerCommand())
	rootCmd.AddCommand(commands.FileCommand())
	rootCmd.AddCommand(commands.AccountCommand())

	err := rootCmd.Execute()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
