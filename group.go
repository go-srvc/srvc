package srvc

import (
	"errors"
	"sync"
)

// JoinErrors is used by srvc to combine multiple errors into one.
// Override it to plug in custom multi-error formatting.
var JoinErrors = errors.Join

// ErrStr adds Error method to string type.
type ErrStr string

func (e ErrStr) Error() string { return string(e) }

// ErrGroup is a goroutine group that waits for all goroutines to finish and collects errors.
type ErrGroup struct {
	wg    sync.WaitGroup
	mutex sync.RWMutex
	err   error
}

// Go runs the given function in a goroutine.
func (eg *ErrGroup) Go(f func() error) {
	eg.wg.Add(1)

	go func() {
		defer eg.wg.Done()
		if err := f(); err != nil {
			eg.mutex.Lock()
			eg.err = JoinErrors(eg.err, err)
			eg.mutex.Unlock()
		}
	}()
}

// Wait waits for all goroutines to finish and returns all errors that occurred.
func (eg *ErrGroup) Wait() error {
	eg.wg.Wait()
	eg.mutex.RLock()
	err := eg.err
	eg.mutex.RUnlock()
	return err
}
