package main

import (
	"os"

	"json2table/internal/app"
	"json2table/internal/cli"
	"json2table/internal/desktopgui"
	"json2table/internal/webui"
)

func main() {
	program := app.App{
		RunCLI:   cli.Run,
		RunGUI:   desktopgui.Run,
		RunWebUI: webui.Run,
	}
	os.Exit(program.Run(os.Args[1:], os.Stdout, os.Stderr))
}
