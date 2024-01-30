package commands

import (
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

type serverFlags struct {
	address   string
	engine    string
	key       string
	secret    string
	data      string
	readyChan chan bool
}

func runServer(flags serverFlags) {
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

	mux := flow.New()
	mux.Use(httplog.RequestLogger(httplog.NewLogger("sc", httplog.Options{
		LogLevel:         slog.LevelInfo,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
		TimeFieldFormat:  time.DateTime,
		Tags: map[string]string{
			"version": "v1.0-81aa4244d9fc8076a",
			"env":     "dev",
		},
		QuietDownRoutes: []string{
			"/",
			"/ping",
		},
		QuietDownPeriod: 10 * time.Second,
		// SourceFieldName: "source",
	}),
	))

	mux.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			slog.Info("R:", "remote", r.RemoteAddr, "proto", r.Proto, "method", r.Method, "path", r.URL.RequestURI(), slog.Duration("duration", time.Since(start)))

			handler.ServeHTTP(w, r)

		})
	})

	eng, err := engine.New()
	if err != nil {
		slog.Error("could not create engine")
		os.Exit(2)
	}

	v22.Setup(mux, eng, flags.key, flags.secret, flags.data)

	srv := &http.Server{
		Addr:         flags.address,
		Handler:      mux,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}
	slog.Debug("storage directory", "path", flags.data)

	fmt.Println("----")
	fmt.Println("Scanii mock server is starting...")
	fmt.Println("→ using API Key:", flags.key)
	fmt.Println("→ using API Secret:", flags.secret)
	fmt.Printf("→ engines with %d known rules\n", eng.RuleCount())
	//goland:noinspection HttpUrlsUsage
	fmt.Printf("✔️ Mock server started on http://%s 🚀\n", flags.address)
	//goland:noinspection HttpUrlsUsage
	fmt.Printf("sample usage: curl -u %s:%s http://%s/v2.2/ping\n", flags.key, flags.secret, flags.address)

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
			runServer(serverF)
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
