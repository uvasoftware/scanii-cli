package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"github.com/alexedwards/flow"
	"github.com/google/gops/agent"
	"log/slog"
	"net/http"
	"os"
	"scanii-cli/internal/engine"
	"scanii-cli/internal/identifiers"
	v22 "scanii-cli/internal/v22"
	"time"
)

type serverFlags struct {
	address string
	engine  string
	key     string
	secret  string
}

func runServer(flags serverFlags) {
	if err := agent.Listen(agent.Options{}); err != nil {
		slog.Error("Failed to start gops agent", "error", err)
		os.Exit(2)
	}

	if flags.key == "" {
		fmt.Println("No API key provided, generating one...")
		flags.key = fmt.Sprintf("akk_%s", identifiers.GenerateShort())
	}

	if flags.secret == "" {
		fmt.Println("No API secret provided, generating one...")
		flags.secret = fmt.Sprintf("aks_%s", identifiers.GenerateSecure())
	}

	//goland:noinspection HttpUrlsUsage
	fmt.Println("→ API Key:", flags.key)
	fmt.Println("→ API Secret:", flags.secret)
	fmt.Println("")
	//goland:noinspection HttpUrlsUsage
	fmt.Printf("Started mock server on http://%s 🚀\n", flags.address)

	mux := flow.New()

	mux.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			handler.ServeHTTP(w, r)
			slog.Info("R:", "remote", r.RemoteAddr, "proto", r.Proto, "method", r.Method, "path", r.URL.RequestURI(), slog.Duration("duration", time.Since(start)))

		})
	})

	//mux.Use(middleware.OapiRequestValidator(swagger))
	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if ok {
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))
				expectedUsernameHash := sha256.Sum256([]byte(flags.key))
				expectedPasswordHash := sha256.Sum256([]byte(flags.secret))

				usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
				passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

				if usernameMatch && passwordMatch {
					next.ServeHTTP(w, r)
					return
				}
			}

			err := writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Apologies but we could not authenticate this request.",
			}, http.Header{"WWW-Authenticate": {" Basic realm=Scanii API"}})
			if err != nil {
				slog.Error(err.Error())
			}
		})
	})

	eng, err := engine.New()
	if err != nil {
		slog.Error("could not create engine")
		os.Exit(2)
	}

	v22.Setup(mux, eng, flags.key, flags.secret)

	srv := &http.Server{
		Addr:         flags.address,
		Handler:      mux,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
