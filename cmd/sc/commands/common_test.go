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
	endpoint = fmt.Sprintf("localhost:%d", 20_000+rand.Intn(1000))
	ready := make(chan bool)
	go func() {
		runServer(serverFlags{
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
		ApiKey:    key,
		ApiSecret: secret,
		Endpoint:  endpoint,
	}
}
