// Package srvc provides simple but powerful Run functionality on top of Module abstraction.
// Ready made modules can be found under: github.com/go-srvc/mods.
package srvc

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
)

const ErrModulePanic = ErrStr("module recovered from panic")

// exitFn allows patching the os.Exit function for testing purposes.
var exitFn = os.Exit

type Module interface {
	// ID should return identifier for logging purposes.
	ID() string
	// Init allows synchronous initialization of module.
	Init() error
	// Run should start the module and block until stop is called or error occurs.
	Run() error
	// Stop allows synchronous cleanup of module should make Run() return eventually.
	// If Init was called then Stop is guaranteed to be called as part of cleanup.
	Stop() error
}

// Run will run all given modules using following control flow:
//
//  1. Exec Init() for each module in order.
//     If any Init() returns error the Init loop is stopped and Stop() will be called
//     for already initialized modules in reverse order.
//  2. Exec Run() for each module in own goroutine so order isn't guaranteed.
//  3. Wait for any Run() function to return nil or an error and move to Stop loop.
//  4. Exec Stop() for modules in reverse order.
//  5. Wait for all Run() goroutines to return.
//  6. Return all errors or nil
//
// Possible panics inside modules are captured to allow graceful shutdown of other modules.
// Captured panics are converted into errors and ErrPanic is returned.
func Run(modules ...Module) error {
	slog.Info("starting service")
	if err := run(modules...); err != nil {
		slog.Error("service exited with error", slog.Any("error", err))
		return err
	}
	slog.Info("service exited successfully")
	return nil
}

// RunAndExit is convenience wrapper for Run that calls os.Exit with code 1 in case of an error.
// The common use case is to srvc.RunAndExit from main function and let the srvc handle the rest.
//
//	package main
//
//	import "github.com/go-srvc/srvc"
//
//	func main() {
//		srvc.RunAndExit(
//		// Add your modules here
//		)
//	}
func RunAndExit(modules ...Module) {
	if err := Run(modules...); err != nil {
		exitFn(1)
	}
}

func run(modules ...Module) error {
	if len(modules) == 0 {
		slog.Warn("no modules to run")
		return nil
	}

	wg := &ErrGroup{}
	initialized, initErr := initialize(modules...)
	if initErr == nil {
		execute(wg, initialized...)
	}

	slog.Info("stopping modules")
	var stopErr error
	for i := len(initialized) - 1; i >= 0; i-- {
		mod := initialized[i]
		slog.Info("module stopping", slog.String("name", mod.ID()))
		stopErr = JoinErrors(stopErr, catchPanic(mod.Stop))
		slog.Info("module stopped", slog.String("name", mod.ID()))
	}

	return JoinErrors(initErr, wg.Wait(), stopErr)
}

func initialize(modules ...Module) ([]Module, error) {
	slog.Info("initializing modules")
	initialized := make([]Module, 0, len(modules))
	for _, mod := range modules {
		slog.Info("module initializing", slog.String("name", mod.ID()))
		err := catchPanic(mod.Init)
		if err != nil {
			return initialized, fmt.Errorf("failed to initialize module %s: %w", mod.ID(), err)
		}
		initialized = append(initialized, mod)
		slog.Info("module initialized", slog.String("name", mod.ID()))
	}
	slog.Info("all modules initialized successfully")
	return initialized, nil
}

func execute(wg *ErrGroup, modules ...Module) {
	slog.Info("starting modules")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, mod := range modules {
		wg.Go(func() error {
			defer func() {
				slog.Info("module exited", slog.String("name", mod.ID()))
				cancel()
			}()

			slog.Info("module started", slog.String("name", mod.ID()))
			err := catchPanic(mod.Run)
			if err != nil {
				return fmt.Errorf("failed to run module %s: %w", mod.ID(), err)
			}

			return nil
		})
	}

	<-ctx.Done()
}

func catchPanic(fn func() error) (err error) {
	defer func() {
		if rErr := recover(); rErr != nil {
			// Print stack trace to log without logger to preserver proper multiline formatting.
			fmt.Println(string(debug.Stack()))
			err = fmt.Errorf("%w: %s", ErrModulePanic, rErr)
		}
	}()
	return fn()
}
