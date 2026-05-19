package srvc_test

import (
	"fmt"

	"github.com/go-srvc/srvc"
)

type printMod struct{}

func (m *printMod) ID() string  { return "printMod" }
func (m *printMod) Init() error { return nil }
func (m *printMod) Run() error  { fmt.Println("hello"); return nil }
func (m *printMod) Stop() error { return nil }

func ExampleRun() {
	_ = srvc.Run(&printMod{})
}
