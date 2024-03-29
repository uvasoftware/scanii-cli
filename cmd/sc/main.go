package main

import (
	"context"
	"fmt"
	"github.com/google/gops/agent"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/uvasoftware/scanii-cli/cmd/sc/internal/commands"
	"log/slog"
	"os"
	"runtime/debug"
)

var (
	verbose bool

	// These variables are set in the build step
	version = "dev"     //nolint
	date    = "unknown" //nolint
)

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
		Run: func(_ *cobra.Command, _ []string) {
			bi, _ := debug.ReadBuildInfo()
			fmt.Println("------------------------------------------------------------")
			fmt.Printf("%-15s: %s\n", "Version", version)
			fmt.Printf("%-15s: %s\n", "Date", date)
			fmt.Printf("%-15s: %s\n", "Go Version", bi.GoVersion)
			fmt.Println("------------------------------------------------------------")
			fmt.Println("Build settings:")
			for _, e := range bi.Settings {
				fmt.Printf("  %-15s: %s\n", e.Key, e.Value)
			}
		},
	}

	ctx := context.Background()

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(commands.PingCommand(ctx))
	rootCmd.AddCommand(commands.FileCommand(ctx))
	rootCmd.AddCommand(commands.AccountCommand(ctx))
	rootCmd.AddCommand(commands.AuthTokenCommand(ctx))
	rootCmd.AddCommand(commands.ServerCommand())
	rootCmd.AddCommand(commands.ConfigureCommand())
	rootCmd.AddCommand(versionCmd)

	err := rootCmd.Execute()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
