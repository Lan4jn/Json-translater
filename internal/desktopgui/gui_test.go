package desktopgui

import "testing"

func TestDefaultOutputPathUsesCSVNextToInput(t *testing.T) {
	got := defaultOutputPath(`C:\data\people.json`)
	want := `C:\data\people.csv`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestDefaultOutputPathForFormatUsesXLSX(t *testing.T) {
	got := defaultOutputPathForFormat(`C:\data\people.json`, "xlsx")
	want := `C:\data\people.xlsx`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
