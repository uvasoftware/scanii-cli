package main

import (
	"fmt"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var verbose bool

func init() {

}

func main() {

	rootCmd := &cobra.Command{
		Use:           "sc",
		SilenceUsage:  true,
		SilenceErrors: true,
		Annotations: map[string]string{
			"get":       "http",
			"post":      "http",
			"delete":    "http",
			"trigger":   "webhooks",
			"listen":    "webhooks",
			"logs":      "stripe",
			"status":    "stripe",
			"resources": "resources",
		},
		Version: "0.0.1",
		Short:   "A CLI to help you integrate Scanii with your application",
		Long:    "",
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
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Hugo",
		Long:  `All software has versions. This is Hugo's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello version")
		},
	}

	serverF := serverFlags{}
	serverCmd := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(serverF)
		},
		Short: "Starts a mock server suitable for testing purposes",
	}

	serverCmd.PersistentFlags().StringVar(&serverF.address, "address", "localhost:4000", "Address to listen on")
	serverCmd.PersistentFlags().StringVar(&serverF.engine, "engine", "", "Optional engine config to load")
	serverCmd.PersistentFlags().StringVar(&serverF.key, "key", "akk_dDCetnjoWQSVtns2", "API key to use, if not provided will be dynamically generated")
	serverCmd.PersistentFlags().StringVar(&serverF.secret, "secret", "aks_wayY13ZZlsLswr0hA6N6Wp3BtEi6YPR6", "API secret to use, if not provided will be dynamically generated")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serverCmd)

	err := rootCmd.Execute()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	//subcommands.Register(subcommands.HelpCommand(), "")
	//subcommands.Register(subcommands.FlagsCommand(), "")
	//subcommands.Register(subcommands.CommandsCommand(), "")
	//subcommands.Register(&serverCommand{}, "")
	//flag.Parse()
	//ctx := context.Background()
	//os.Exit(int(subcommands.Execute(ctx)))
}
