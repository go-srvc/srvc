package srvc

import "sync"

// InitMod wraps initFn into a Module that calls initFn during the init phase.
// Run blocks until the service shuts down and Stop is a no-op.
// Useful for one time setup that has no long running work of its own.
func InitMod(id string, initFn func() error) Module {
	return &fnMod{id: id, init: initFn, done: make(chan struct{})}
}

// RunMod wraps runFn into a Module that calls runFn during the run phase.
// Init and Stop are no-ops, so runFn must return on its own.
// When it returns the whole service moves to the stop sequence.
func RunMod(id string, runFn func() error) Module {
	return &fnMod{id: id, run: runFn, done: make(chan struct{})}
}

// StopMod wraps stopFn into a Module that calls stopFn during the stop phase.
// Init is a no-op and Run blocks until the service shuts down.
// Useful for cleanup that should always run when the service exits.
func StopMod(id string, stopFn func() error) Module {
	return &fnMod{id: id, stop: stopFn, done: make(chan struct{})}
}

// fnMod implements Module for a single user given function.
// Nil functions are no-ops and a nil run blocks until Stop is called.
type fnMod struct {
	id   string
	init func() error
	run  func() error
	stop func() error
	done chan struct{}
	once sync.Once
}

func (m *fnMod) ID() string { return m.id }

func (m *fnMod) Init() error {
	if m.init == nil {
		return nil
	}
	return m.init()
}

func (m *fnMod) Run() error {
	if m.run == nil {
		<-m.done
		return nil
	}
	return m.run()
}

func (m *fnMod) Stop() error {
	defer m.once.Do(func() { close(m.done) })
	if m.stop == nil {
		return nil
	}
	return m.stop()
}
