//go:build windows

package desktopgui

import "testing"

func TestSignedInt32ParamPassesNegativeFontHeight(t *testing.T) {
	got := signedInt32Param(-15)
	want := ^uintptr(14)
	if got != want {
		t.Fatalf("expected %#x, got %#x", want, got)
	}
}
