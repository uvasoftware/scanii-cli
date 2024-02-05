//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2 --config=oapi.yaml v22.yaml
package v22

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"github.com/alexedwards/flow"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"scanii-cli/internal/engine"
	"scanii-cli/internal/identifiers"
	"strings"
)

func Setup(mux *flow.Mux, eng *engine.Engine, key, secret, data string, baseUrl string) {
	handlers := FakeHandler{
		engine:  eng,
		store:   store{path: data},
		baseurl: baseUrl,
	}

	swagger, err := GetSwagger()
	if err != nil {
		panic(err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	mux.Group(func(router *flow.Mux) {

		//router.Use(middleware.OapiRequestValidator(swagger))
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username, password, ok := r.BasicAuth()
				if ok {

					// we allow two forms of authentication, API keys and auth tokens
					if password == "" {
						token := &AuthToken{}
						err = handlers.store.load(username, token)
						if err != nil {
							slog.Error("failed to load token", "error", err)
						} else {
							ctxWithKey := context.WithValue(r.Context(), "key", token.Id)
							ctx := context.WithValue(ctxWithKey, "key", token.Id)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}

					} else {
						usernameHash := sha256.Sum256([]byte(username))
						passwordHash := sha256.Sum256([]byte(password))
						expectedUsernameHash := sha256.Sum256([]byte(key))
						expectedPasswordHash := sha256.Sum256([]byte(secret))

						usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
						passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

						if usernameMatch && passwordMatch {
							ctxWithKey := context.WithValue(r.Context(), "key", username)
							ctx := context.WithValue(ctxWithKey, "key", username)

							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}

				err := writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "Apologies but we could not authenticate this request.",
				}, http.Header{"WWW-Authenticate": {" Basic realm=Scanii API"}})
				if err != nil {
					panic(err)
				}
			})
		})
		router.Use(func(next http.Handler) http.Handler {
			hostId := "hst_" + identifiers.GenerateShort()
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("XScaniiRequestId", "req_"+identifiers.GenerateShort())
				w.Header().Set("XScaniiHostId", hostId)
				next.ServeHTTP(w, r)
			})

		})

		router.HandleFunc("/v2.2/account.json", handlers.Account, "GET")
		router.HandleFunc("/v2.2/ping", handlers.Ping, "GET")
		router.HandleFunc("/v2.2/files/async", handlers.ProcessFileAsync, "POST")
		router.HandleFunc("/v2.2/files/fetch", handlers.ProcessFileFetch, "POST")
		router.HandleFunc("/v2.2/files", handlers.ProcessFile, "POST")
		router.HandleFunc("/v2.2/files/:id", func(writer http.ResponseWriter, request *http.Request) {
			handlers.RetrieveFile(writer, request, flow.Param(request.Context(), "id"))
		}, "GET")

		router.HandleFunc("/v2.2/auth/tokens", handlers.CreateToken, "POST")
		router.HandleFunc("/v2.2/auth/tokens/:id", func(writer http.ResponseWriter, request *http.Request) {
			handlers.RetrieveToken(writer, request, flow.Param(request.Context(), "id"))
		}, "GET")
		router.HandleFunc("/v2.2/auth/tokens/:id", func(writer http.ResponseWriter, request *http.Request) {
			handlers.DeleteToken(writer, request, flow.Param(request.Context(), "id"))
		}, "DELETE")
	})

}

func generateId() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
