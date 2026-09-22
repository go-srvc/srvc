[![Go Reference](https://pkg.go.dev/badge/github.com/go-srvc/srvc.svg)](https://pkg.go.dev/github.com/go-srvc/srvc) [![codecov](https://codecov.io/github/go-srvc/srvc/graph/badge.svg?token=H3u7Ui9PfC)](https://codecov.io/github/go-srvc/srvc) ![main](https://github.com/go-srvc/srvc/actions/workflows/go.yaml/badge.svg?branch=main)

# Simple, Safe, and Modular Service Runner

srvc library provides a simple but powerful interface with zero external dependencies for running service modules.

The library itself stays dependency free: it ships the runner, the `Module` interface, and a handful of thin adapters (`InitMod`, `RunMod`, `StopMod`, `CtxMod`, `IdleMod`) that turn plain functions into modules. They are glue, nothing more. Modules that wrap real dependencies and carry their own domain logic, such as HTTP servers, signal handling, databases, and telemetry, live in [github.com/go-srvc/mods](https://github.com/go-srvc/mods).

## Use Case

Normally Go services are composed of multiple "modules" which each run in their own goroutine such as http server, signal listener, kafka consumer, ticker, etc. These modules should remain alive throughout the lifecycle of the whole service, and if one goes down, graceful exit should be executed to avoid "zombie" services. srvc takes care of all this via a simple module interface.

Ready made modules with batteries included can be found under [github.com/go-srvc/mods](https://github.com/go-srvc/mods), while this repo keeps only the runner and the function adapters.

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

### Function modules

`InitMod`, `RunMod`, `StopMod`, `CtxMod`, and `IdleMod` wrap plain functions into a module when the full interface is not needed. They are deliberately minimal adapters with no logic of their own; anything that needs configuration, defaults, or dependency specific behaviour belongs in a real module, either your own or one from [mods](https://github.com/go-srvc/mods).

```go
package main

import "github.com/go-srvc/srvc"

func main() {
 srvc.RunAndExit(
  // Called during init, Run blocks until shutdown.
  srvc.InitMod("migrations", runMigrations),
  // Called during stop, in reverse order like any other module.
  srvc.StopMod("flush", flushBuffers),
  // Called during run, service shuts down once it returns.
  srvc.RunMod("job", doWork),
  // Called during run with a context that is canceled on shutdown.
  srvc.CtxMod("consumer", consume),
  // Init and stop pair without any work of its own during run.
  srvc.IdleMod("db", openDB, closeDB),
 )
}
```

`RunMod` and `CtxMod` have no-op `Init`, and their function must return on its own for the service to shut down. `CtxMod` cancels the context passed to its function when `Stop` is called, so a function that respects the context returns on shutdown. `InitMod`, `StopMod`, and `IdleMod` have no work of their own during run, so their `Run` blocks until the service shuts down.

## Lifecycle

`Run` executes modules through a deterministic lifecycle:

1. **Init** is called sequentially in the order modules are passed. If any `Init` returns an error, the loop stops and `Stop` is called on already-initialized modules in reverse order. Uninitialized modules never get `Init` *or* `Stop`.
2. **Run** is started for each successfully initialized module in its own goroutine. Start order is not guaranteed.
3. When the first `Run` returns (with or without error), the service moves to shutdown.
4. **Stop** is called sequentially in **reverse** order on every initialized module. Each module's `Stop` must cause its `Run` to return.
5. `Run` blocks until every `Run` goroutine has returned, then returns the joined errors from `Init`, `Run`, and `Stop` (or `nil`).

### Panic recovery

Panics inside `Init`, `Run`, or `Stop` are recovered. The stack trace is logged, and the panic is converted into an error wrapping `srvc.ErrModulePanic` so other modules can still shut down gracefully.

### Exit behaviour

`RunAndExit` calls `os.Exit(1)` if `Run` returns any error, and returns normally on success.

### Contracts modules must uphold

- `Stop` must make `Run` return. If `Run` ignores `Stop`, `srvc.Run` will block in its final wait. There is no built-in shutdown timeout, so a stuck module hangs the service.
- `ID` should return a stable, unique identifier used for log attribution.

### Why no context.Context on lifecycle methods?

`Init`, `Run`, and `Stop` deliberately take no `context.Context`. Shutdown deadlines are already enforced by the platform the service runs on (Kubernetes `terminationGracePeriodSeconds`, systemd `TimeoutStopSec`, ECS `stopTimeout`), and a Go-side deadline cannot extend past the platform `SIGKILL`. Adding context plumbing to the interface would expand the surface without giving modules any extra safety. Modules that need a startup or shutdown budget for their own internal calls can take a context via their own options without changing `Module`. Functions that already take a context can be wrapped with `CtxMod`, which keeps the plumbing in the helper instead of the interface.

## Acknowledgements

This library is something I have found myself writing over and over again in every project I been part of. One of the iterations can be found under [https://github.com/elisasre/go-common](https://github.com/elisasre/go-common).
