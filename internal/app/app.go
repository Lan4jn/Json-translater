package app

import (
	"fmt"
	"io"
)

type anyWriter = io.Writer

type App struct {
	RunCLI   func(args []string, stdout, stderr io.Writer) int
	RunGUI   func() error
	RunWebUI func() error
}

func (a App) Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--webui" || args[0] == "-webui" || args[0] == "webui") {
		if err := a.RunWebUI(); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	if len(args) > 0 {
		return a.RunCLI(args, stdout, stderr)
	}
	if err := a.RunGUI(); err != nil {
		if a.RunWebUI == nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stderr, "桌面 GUI 启动失败，已切换到 WebUI: %v\n", err)
		if webErr := a.RunWebUI(); webErr != nil {
			fmt.Fprintln(stderr, webErr)
			return 1
		}
	}
	return 0
}
