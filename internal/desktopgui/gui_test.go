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

func TestSelectedPathFromPortalURIsDecodesFileURI(t *testing.T) {
	got, err := selectedPathFromPortalURIs([]string{"file:///home/uos/%E6%95%B0%E6%8D%AE.json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/home/uos/数据.json"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSelectedPathFromPortalURIsRejectsNonFileURI(t *testing.T) {
	if _, err := selectedPathFromPortalURIs([]string{"https://example.com/data.json"}); err == nil {
		t.Fatal("expected error for non-file URI")
	}
}
