package commands

import (
	"fmt"
	"math/rand"
	"time"
)

var config *configuration

const (
	key    = "key"
	secret = "secret"
)

var (
	endpoint = ""
)

func init() {
	endpoint = fmt.Sprintf("localhost:%d", 20_000+rand.Intn(1000)) //nolint:gosec
	ready := make(chan bool)
	go func() {
		runServer(&serverFlags{
			key:       key,
			secret:    secret,
			address:   endpoint,
			readyChan: ready,
		})
	}()

	// waiting for server to start
	<-ready
	config = &configuration{
		Updated:   time.Now(),
		APIKey:    key,
		APISecret: secret,
		Endpoint:  endpoint,
	}
}
