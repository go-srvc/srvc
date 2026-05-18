package srvc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-srvc/srvc"
)

const (
	errInit = srvc.ErrStr("init error")
	errRun  = srvc.ErrStr("run error")
	errStop = srvc.ErrStr("stop error")
)

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		mods        []srvc.Module
		expectedErr error
	}{
		{
			name:        "NoError",
			mods:        nil,
			expectedErr: nil,
		},
		{
			name:        "NoError_MultipleModules",
			mods:        []srvc.Module{SuccessMod(), SuccessMod(), SuccessMod()},
			expectedErr: nil,
		},
		{
			name:        "InitError",
			mods:        []srvc.Module{InitErrMod()},
			expectedErr: errInit,
		},
		{
			name:        "InitError_and_StopError_Different_Modules",
			mods:        []srvc.Module{StopErrMod(), InitErrMod()},
			expectedErr: errStop,
		},
		{
			name:        "RunError",
			mods:        []srvc.Module{RunErrMod()},
			expectedErr: errRun,
		},
		{
			name:        "StopError",
			mods:        []srvc.Module{StopErrMod()},
			expectedErr: errStop,
		},
		{
			name:        "InitPanic",
			mods:        []srvc.Module{InitPanicMod()},
			expectedErr: srvc.ErrModulePanic,
		},
		{
			name:        "RunPanic",
			mods:        []srvc.Module{RunPanicMod()},
			expectedErr: srvc.ErrModulePanic,
		},
		{
			name:        "StopPanic",
			mods:        []srvc.Module{StopPanicMod()},
			expectedErr: srvc.ErrModulePanic,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stopMod, stop := StopMod()
			tc.mods = append(tc.mods, stopMod)

			go func() {
				time.Sleep(time.Second)
				stop()
			}()

			err := srvc.Run(tc.mods...)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected to found error %v from %v", tc.expectedErr, err)
			}
		})
	}
}

func TestRun_NoModules(t *testing.T) {
	err := srvc.Run()
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

func TestRun_StopTimeout(t *testing.T) {
	srvc.StopTimeout = 50 * time.Millisecond
	t.Cleanup(func() { srvc.StopTimeout = 0 })

	hang := make(chan struct{})
	t.Cleanup(func() { close(hang) })

	mod := &TestMod{
		init: func() error { return nil },
		run:  func() error { <-hang; return nil },
		stop: func() error { <-hang; return nil },
	}

	stopMod, stop := StopMod()
	go func() {
		time.Sleep(10 * time.Millisecond)
		stop()
	}()

	err := srvc.Run(mod, stopMod)
	if !errors.Is(err, srvc.ErrStopTimeout) {
		t.Errorf("expected ErrStopTimeout, got %v", err)
	}
}

// StopMod can be used to trigger stop sequence for srvc.Run in tests.
func StopMod() (srvc.Module, func()) {
	ctx, stop := context.WithCancel(context.Background())
	return &TestMod{
		init: func() error {
			return nil
		},
		run: func() error {
			<-ctx.Done()
			return nil
		},
		stop: func() error {
			stop()
			return nil
		},
	}, stop
}

func InitErrMod() srvc.Module {
	return &TestMod{
		init: func() error { return errInit },
		run:  func() error { return nil },
		stop: func() error { return nil },
	}
}

func InitPanicMod() srvc.Module {
	return &TestMod{
		run:  func() error { return nil },
		stop: func() error { return nil },
	}
}

func RunErrMod() srvc.Module {
	return &TestMod{
		init: func() error { return nil },
		run:  func() error { return errRun },
		stop: func() error { return nil },
	}
}

func RunPanicMod() srvc.Module {
	return &TestMod{
		init: func() error { return nil },
		stop: func() error { return nil },
	}
}

func StopErrMod() srvc.Module {
	return &TestMod{
		init: func() error { return nil },
		run:  func() error { return nil },
		stop: func() error { return errStop },
	}
}

func InitStopErrMod() srvc.Module {
	return &TestMod{
		init: func() error { return errInit },
		run:  func() error { return nil },
		stop: func() error { return errStop },
	}
}

func StopPanicMod() srvc.Module {
	return &TestMod{
		init: func() error { return nil },
		run:  func() error { return nil },
	}
}

func SuccessMod() srvc.Module {
	mod, _ := StopMod()
	return mod
}

type TestMod struct{ init, run, stop func() error }

func (m *TestMod) ID() string  { return "TestMod" }
func (m *TestMod) Init() error { return m.init() }
func (m *TestMod) Run() error  { return m.run() }
func (m *TestMod) Stop() error { return m.stop() }
