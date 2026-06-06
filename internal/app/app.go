package app

import (
	"fmt"
	"io"
)

type anyWriter = io.Writer

type App struct {
	RunCLI func(args []string, stdout, stderr io.Writer) int
	RunGUI func() error
}

func (a App) Run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		return a.RunCLI(args, stdout, stderr)
	}
	if err := a.RunGUI(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
