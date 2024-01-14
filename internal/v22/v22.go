package v22

import (
	"github.com/alexedwards/flow"
	"scanii-cli/internal/engine"
)

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2 --config=oapi.yaml v22.yaml

func Setup(mux *flow.Mux, eng *engine.Engine, key, secret string) {
	si := NewStrictHandler(MockHandler{
		engine: eng,
	}, nil)

	swagger, err := GetSwagger()
	if err != nil {
		panic(err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	mux.Group(func(mux *flow.Mux) {
		mux.HandleFunc("/v2.2/account.json", si.Account, "GET")
		mux.HandleFunc("/v2.2/ping", si.Ping, "GET")
		mux.HandleFunc("/v2.2/files", si.ProcessFile, "POST")
	})

}
