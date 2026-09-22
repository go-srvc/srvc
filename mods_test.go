package srvc_test

import (
	"context"
	"testing"

	"github.com/go-srvc/srvc"
)

func TestInitMod(t *testing.T) {
	called := false
	mod := srvc.InitMod("initMod", func() error { called = true; return nil })

	equal(t, "initMod", mod.ID())
	noError(t, mod.Init())
	if !called {
		t.Error("initFn was not called")
	}

	go func() { noError(t, mod.Stop()) }()
	noError(t, mod.Run())
}

func TestInitMod_Error(t *testing.T) {
	mod := srvc.InitMod("initMod", func() error { return errInit })
	errorIs(t, srvc.Run(mod), errInit)
}

func TestRunMod(t *testing.T) {
	called := false
	mod := srvc.RunMod("runMod", func() error { called = true; return nil })

	equal(t, "runMod", mod.ID())
	noError(t, srvc.Run(mod))
	if !called {
		t.Error("runFn was not called")
	}
}

func TestRunMod_Error(t *testing.T) {
	mod := srvc.RunMod("runMod", func() error { return errRun })
	errorIs(t, srvc.Run(mod), errRun)
}

func TestStopMod(t *testing.T) {
	called := false
	mod := srvc.StopMod("stopMod", func() error { called = true; return nil })

	equal(t, "stopMod", mod.ID())
	noError(t, srvc.Run(mod, srvc.RunMod("trigger", func() error { return nil })))
	if !called {
		t.Error("stopFn was not called")
	}
}

func TestStopMod_Error(t *testing.T) {
	mod := srvc.StopMod("stopMod", func() error { return errStop })
	errorIs(t, srvc.Run(mod, srvc.RunMod("trigger", func() error { return nil })), errStop)
}

func TestMods_Order(t *testing.T) {
	order := []string{}
	err := srvc.Run(
		srvc.InitMod("init", func() error { order = append(order, "init"); return nil }),
		srvc.StopMod("stop", func() error { order = append(order, "stop"); return nil }),
		srvc.RunMod("run", func() error { order = append(order, "run"); return nil }),
	)
	noError(t, err)

	equal(t, 3, len(order))
	equal(t, "init", order[0])
	equal(t, "run", order[1])
	equal(t, "stop", order[2])
}

func TestCtxMod(t *testing.T) {
	var ctxErr error
	mod := srvc.CtxMod("ctxMod", func(ctx context.Context) error {
		<-ctx.Done()
		ctxErr = ctx.Err()
		return nil
	})

	equal(t, "ctxMod", mod.ID())
	noError(t, srvc.Run(mod, srvc.RunMod("trigger", func() error { return nil })))
	errorIs(t, ctxErr, context.Canceled)
}

func TestCtxMod_Error(t *testing.T) {
	mod := srvc.CtxMod("ctxMod", func(ctx context.Context) error { return errRun })
	errorIs(t, srvc.Run(mod), errRun)
}

func TestIdleMod(t *testing.T) {
	order := []string{}
	mod := srvc.IdleMod("idleMod",
		func() error { order = append(order, "init"); return nil },
		func() error { order = append(order, "stop"); return nil },
	)

	equal(t, "idleMod", mod.ID())
	noError(t, srvc.Run(mod, srvc.RunMod("trigger", func() error {
		order = append(order, "run")
		return nil
	})))

	equal(t, 3, len(order))
	equal(t, "init", order[0])
	equal(t, "run", order[1])
	equal(t, "stop", order[2])
}

func TestIdleMod_NilFns(t *testing.T) {
	mod := srvc.IdleMod("idleMod", nil, nil)
	noError(t, srvc.Run(mod, srvc.RunMod("trigger", func() error { return nil })))
}

func TestIdleMod_InitError(t *testing.T) {
	mod := srvc.IdleMod("idleMod", func() error { return errInit }, nil)
	errorIs(t, srvc.Run(mod), errInit)
}

func TestIdleMod_StopError(t *testing.T) {
	mod := srvc.IdleMod("idleMod", nil, func() error { return errStop })
	errorIs(t, srvc.Run(mod, srvc.RunMod("trigger", func() error { return nil })), errStop)
}
