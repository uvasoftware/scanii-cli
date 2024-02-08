package commands

import (
	"embed"
	"errors"
	"fmt"
	"github.com/alexedwards/flow"
	"github.com/go-chi/httplog/v2"
	"github.com/spf13/cobra"
	"log/slog"
	"net"
	"net/http"
	"os"
	"scanii-cli/internal/engine"
	"scanii-cli/internal/identifiers"
	v22 "scanii-cli/internal/v22"
	"time"
)

//go:embed  static/**
var static embed.FS

type serverFlags struct {
	address   string
	engine    string
	key       string
	secret    string
	data      string
	readyChan chan bool
}

func runServer(flags *serverFlags) {
	if flags.key == "" {
		fmt.Println("No API key provided, generating one...")
		flags.key = fmt.Sprintf("akk_%s", identifiers.GenerateShort())
	}

	if flags.secret == "" {
		fmt.Println("No API secret provided, generating one...")
		flags.secret = fmt.Sprintf("aks_%s", identifiers.GenerateSecure())
	}

	if flags.data == "" {
		dir, err := os.MkdirTemp("", "scanii-cli")

		if err != nil {
			panic(err)
		}
		flags.data = dir
	} else {
		if _, err := os.Stat(flags.data); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(flags.data, 0755)
			if err != nil {
				panic(err)
			}
		}
	}

	router := flow.New()
	router.Use(httplog.RequestLogger(httplog.NewLogger("sc", httplog.Options{
		LogLevel:         slog.LevelInfo,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
		TimeFieldFormat:  time.DateTime,
	}),
	))

	// static static
	fileServer := http.FileServer(http.FS(static))
	router.Handle("/static/...", fileServer, "GET")

	eng, err := engine.New()
	if err != nil {
		slog.Error("could not create engine")
		os.Exit(2)
	}

	v22.Setup(router, eng, flags.key, flags.secret, flags.data, "http://"+flags.address)

	srv := &http.Server{
		Addr:         flags.address,
		Handler:      router,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}
	slog.Debug("storage directory", "path", flags.data)

	fmt.Println("Scanii test server is starting... 🚀")
	fmt.Println("⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻⎻")
	fmt.Println("• Using API Key:", flags.key)
	fmt.Println("• Using API Secret:", flags.secret)
	fmt.Printf("• Engines with %d known rules\n", eng.RuleCount())
	//goland:noinspection HttpUrlsUsage
	fmt.Printf("• Mock server started on http://%s\n", flags.address)
	fmt.Println()
	//goland:noinspection HttpUrlsUsage
	fmt.Printf("Sample usage → curl -u %s:%s http://%s/v2.2/ping\n", flags.key, flags.secret, flags.address)

	listen, err := net.Listen("tcp", flags.address)
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}

	if flags.readyChan != nil {
		flags.readyChan <- true
	}

	err = srv.Serve(listen)
	if err != nil {
		slog.Error("server error", "error", err)
		println(err.Error())
		os.Exit(3)
	}
}

func ServerCommand() *cobra.Command {
	serverF := serverFlags{}
	serverCmd := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(&serverF)
		},
		Short: "Starts a mock server suitable for testing purposes",
	}

	serverCmd.PersistentFlags().StringVarP(&serverF.address, "address", "a", "localhost:4000", "Address to listen on")
	serverCmd.PersistentFlags().StringVarP(&serverF.engine, "engine", "e", "", "Optional engine config to load")
	serverCmd.PersistentFlags().StringVarP(&serverF.data, "data", "d", "", "Result storage path, defaults to a temp directory")
	serverCmd.PersistentFlags().StringVarP(&serverF.key, "key", "k", "key", "API key to use, if not provided will be dynamically generated")
	serverCmd.PersistentFlags().StringVarP(&serverF.secret, "secret", "s", "secret", "API secret to use, if not provided will be dynamically generated")

	return serverCmd
}
