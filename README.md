[![Go Reference](https://pkg.go.dev/badge/github.com/go-srvc/srvc.svg)](https://pkg.go.dev/github.com/go-srvc/srvc) [![codecov](https://codecov.io/github/go-srvc/srvc/graph/badge.svg?token=H3u7Ui9PfC)](https://codecov.io/github/go-srvc/srvc) ![main](https://github.com/go-srvc/srvc/actions/workflows/go.yaml/badge.svg?branch=main)

# Simple, Safe, and Modular Service Runner

srvc library provides a simple but powerful interface with zero external dependencies for running service [modules](https://github.com/go-srvc/mods).

## Use Case

Normally Go services are composed of multiple "modules" which each run in their own goroutine such as http server, signal listener, kafka consumer, ticker, etc. These modules should remain alive throughout the lifecycle of the whole service, and if one goes down, graceful exit should be executed to avoid "zombie" services. srvc takes care of all this via a simple module interface.

List of ready made modules can be found under [github.com/go-srvc/mods](https://github.com/go-srvc/mods)

## Usage

### Main package

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/logmod"
	"github.com/go-srvc/mods/metermod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/mods/sqlxmod"
	"github.com/go-srvc/mods/tracemod"
	"github.com/go-srvc/srvc"
)

func main() {
	db := sqlxmod.New()
	srvc.RunAndExit(
		logmod.New(),
		sigmod.New(),
		tracemod.New(),
		metermod.New(),
		db,
		httpmod.New(
			httpmod.WithAddr(":8080"),
			httpmod.WithHandler(handler(db)),
		),
	)
}

func handler(db *sqlxmod.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := db.DB().PingContext(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "OK")
	})
}
```

### Implementing custom modules

```go
package main

import "github.com/go-srvc/srvc"

func main() {
	srvc.RunAndExit(&MyMod{})
}

type MyMod struct {
	done chan struct{}
}

func (m *MyMod) Init() error {
	m.done = make(chan struct{})
	return nil
}

// Run should block until the module is stopped.
// If you don't have a blocking operation, you can use done channel to block.
func (m *MyMod) Run() error {
	<-m.done
	return nil
}

func (m *MyMod) Stop() error {
	defer close(m.done)
	return nil
}

func (m *MyMod) ID() string { return "MyMod" }
```

## Acknowledgements

This library is something I have found myself writing over and over again in every project I been part of. One of the iterations can be found under [https://github.com/elisasre/go-common](https://github.com/elisasre/go-common).
