package srvc_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-srvc/srvc"
)

func TestErrGroup_MultipleErrors(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	eg := &srvc.ErrGroup{}
	eg.Go(func() error { return err1 })
	eg.Go(func() error { return err2 })
	err := eg.Wait()

	errorIs(t, err, err1)
	errorIs(t, err, err2)
}

func TestErrGroup_Mixed(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	eg := &srvc.ErrGroup{}
	eg.Go(func() error { return nil })
	eg.Go(func() error { return err1 })
	eg.Go(func() error { return nil })
	eg.Go(func() error { return err2 })
	eg.Go(func() error { return nil })
	err := eg.Wait()

	errorIs(t, err, err1)
	errorIs(t, err, err2)
}

func TestErrGroup_NoError(t *testing.T) {
	eg := &srvc.ErrGroup{}
	eg.Go(func() error { return nil })
	eg.Go(func() error { return nil })
	eg.Go(func() error { return nil })
	err := eg.Wait()
	noError(t, err)
}

func TestErrGroup_NoTask(t *testing.T) {
	eg := &srvc.ErrGroup{}
	err := eg.Wait()
	noError(t, err)
}

func TestError_Error(t *testing.T) {
	const (
		msg     = "test error"
		SomeErr = srvc.ErrStr(msg)
	)

	err := fmt.Errorf("more info: %w", SomeErr)
	errorIs(t, err, SomeErr)
	equal(t, msg, errors.Unwrap(err).Error())
}

func errorIs(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("expected to found error %v from %v", target, err)
	}
}

func noError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

func equal[C comparable](t *testing.T, expected, actual C) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}
