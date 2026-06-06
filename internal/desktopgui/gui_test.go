package desktopgui

import "testing"

func TestDefaultOutputPathUsesCSVNextToInput(t *testing.T) {
	got := defaultOutputPath(`C:\data\people.json`)
	want := `C:\data\people.csv`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestPowerShellStringEscapesSingleQuotes(t *testing.T) {
	got := powerShellString(`C:\Users\O'Brien\out.csv`)
	want := `'C:\Users\O''Brien\out.csv'`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
