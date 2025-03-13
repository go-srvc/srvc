package srvc

import (
	"errors"
	"testing"
)

func TestRunAndExit_NoError(t *testing.T) {
	exitFn = func(code int) { t.Errorf("exitFn should not be called but was called with code %d", code) }
	RunAndExit(&TestMod{})
}

func TestRunAndExit_Error(t *testing.T) {
	called := false
	exitFn = func(code int) {
		called = true
		if code != 1 {
			t.Errorf("exitFn was called with wrong code %d", code)
		}
	}

	RunAndExit(&TestMod{err: errors.New("test error")})
	if !called {
		t.Error("exitFn was not called")
	}
}

type TestMod struct {
	err error
}

func (m *TestMod) ID() string  { return "TestMod" }
func (m *TestMod) Init() error { return m.err }
func (m *TestMod) Run() error  { return nil }
func (m *TestMod) Stop() error { return nil }
