package srvc_test

import (
	"fmt"

	"github.com/go-srvc/srvc"
)

func ExampleRun() {
	err := srvc.Run(
	// TODO: Add modules here
	)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ExampleRunAndExit() {
	srvc.RunAndExit(
	// TODO: Add modules here
	)
}
