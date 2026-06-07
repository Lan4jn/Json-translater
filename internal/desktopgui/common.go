package desktopgui

import (
	"path/filepath"
	"strings"
)

func defaultOutputPath(inputPath, format string) string {
	dir := filepath.Dir(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if base == "" || base == "." {
		base = "output"
	}
	ext := ".csv"
	if format == "xlsx" {
		ext = ".xlsx"
	}
	if dir == "." || dir == "" {
		return base + ext
	}
	return filepath.Join(dir, base+ext)
}
