package app

import (
	"bytes"
	"errors"
	"testing"
)

func TestRunWithoutArgsStartsGUI(t *testing.T) {
	calledGUI := false
	calledCLI := false
	app := App{
		RunCLI: func(args []string, stdout, stderr anyWriter) int {
			calledCLI = true
			return 0
		},
		RunGUI: func() error {
			calledGUI = true
			return nil
		},
		RunWebUI: func() error {
			t.Fatal("did not expect WebUI runner to be called")
			return nil
		},
	}

	code := app.Run(nil, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !calledGUI {
		t.Fatal("expected GUI runner to be called")
	}
	if calledCLI {
		t.Fatal("did not expect CLI runner to be called")
	}
}

func TestRunWithArgsStartsCLI(t *testing.T) {
	calledGUI := false
	calledCLI := false
	app := App{
		RunCLI: func(args []string, stdout, stderr anyWriter) int {
			calledCLI = true
			if len(args) != 2 || args[0] != "-input" || args[1] != "data.json" {
				t.Fatalf("unexpected args: %#v", args)
			}
			return 7
		},
		RunGUI: func() error {
			calledGUI = true
			return nil
		},
		RunWebUI: func() error {
			t.Fatal("did not expect WebUI runner to be called")
			return nil
		},
	}

	code := app.Run([]string{"-input", "data.json"}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 7 {
		t.Fatalf("expected CLI exit code 7, got %d", code)
	}
	if !calledCLI {
		t.Fatal("expected CLI runner to be called")
	}
	if calledGUI {
		t.Fatal("did not expect GUI runner to be called")
	}
}

func TestRunWithoutArgsFallsBackToWebUIWhenGUIFails(t *testing.T) {
	calledGUI := false
	calledWebUI := false
	app := App{
		RunCLI: func(args []string, stdout, stderr anyWriter) int {
			t.Fatal("did not expect CLI runner to be called")
			return 0
		},
		RunGUI: func() error {
			calledGUI = true
			return errors.New("Linux 图形界面模式需要安装 zenity、kdialog，或启用 xdg-desktop-portal")
		},
		RunWebUI: func() error {
			calledWebUI = true
			return nil
		},
	}

	code := app.Run(nil, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !calledGUI {
		t.Fatal("expected GUI runner to be called")
	}
	if !calledWebUI {
		t.Fatal("expected WebUI runner to be called")
	}
}

func TestRunWebUIFlagStartsWebUI(t *testing.T) {
	calledWebUI := false
	app := App{
		RunCLI: func(args []string, stdout, stderr anyWriter) int {
			t.Fatal("did not expect CLI runner to be called")
			return 0
		},
		RunGUI: func() error {
			t.Fatal("did not expect GUI runner to be called")
			return nil
		},
		RunWebUI: func() error {
			calledWebUI = true
			return nil
		},
	}

	code := app.Run([]string{"--webui"}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !calledWebUI {
		t.Fatal("expected WebUI runner to be called")
	}
}
