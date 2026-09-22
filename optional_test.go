package srvc_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/go-srvc/srvc"
)

func TestOptional_Enabled(t *testing.T) {
	var initCalled bool
	m := srvc.InitMod("init", func() error { initCalled = true; return nil })
	stop := srvc.RunMod("stop", func() error { return nil })

	if err := srvc.Run(srvc.Optional(true, m), stop); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !initCalled {
		t.Fatal("expected Init to be called")
	}
}

func TestOptional_Disabled(t *testing.T) {
	var called bool
	fn := func() error { called = true; return errors.New("should not be called") }
	m := srvc.IdleMod("disabled", fn, fn)
	stop := srvc.RunMod("stop", func() error { return nil })

	if err := srvc.Run(srvc.Optional(false, m), stop); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected disabled module to be skipped")
	}
}

func TestOptional_Nested(t *testing.T) {
	var called bool
	m := srvc.InitMod("nested", func() error { called = true; return nil })
	stop := srvc.RunMod("stop", func() error { return nil })

	if err := srvc.Run(srvc.Optional(true, srvc.Optional(false, m)), stop); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected nested disabled module to be skipped")
	}
}

func TestOptional_AllDisabled(t *testing.T) {
	m := srvc.InitMod("disabled", func() error { return errors.New("should not be called") })
	if err := srvc.Run(srvc.Optional(false, m)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOptional_EnabledReturnsSameModule(t *testing.T) {
	m := srvc.InitMod("m", nil)
	if got := srvc.Optional(true, m); got != m {
		t.Fatalf("expected enabled module to be returned as is, got: %v", got)
	}
}

func TestOptional_DisabledKeepsID(t *testing.T) {
	m := srvc.Optional(false, srvc.InitMod("metrics", nil))
	if id := m.ID(); id != "metrics" {
		t.Fatalf("expected ID metrics, got: %s", id)
	}
}

func TestOptional_DisabledLogsName(t *testing.T) {
	tests := []struct {
		name     string
		mod      srvc.Module
		expected string
	}{
		{name: "Module", mod: srvc.InitMod("metrics", nil), expected: "name=metrics"},
		{name: "Nil", mod: nil, expected: "name=unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			prev := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
			t.Cleanup(func() { slog.SetDefault(prev) })

			if err := srvc.Run(srvc.Optional(false, tt.mod)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(buf.String(), `msg="module disabled" `+tt.expected) {
				t.Fatalf("expected disabled log with %s, got:\n%s", tt.expected, buf.String())
			}
		})
	}
}
