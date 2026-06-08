package desktopgui

import (
	"path/filepath"
	"strings"

	"json2table/internal/converter"
)

func defaultOutputPath(inputPath string) string {
	return defaultOutputPathForFormat(inputPath, "")
}

func defaultOutputPathForFormat(inputPath, format string) string {
	dir := filepath.Dir(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if base == "" || base == "." {
		base = "output"
	}

	ext := ".csv"
	if format == "xlsx" {
		ext = ".xlsx"
	}
	return filepath.Join(dir, base+ext)
}

func convertFile(inputPath, outputPath, format string) error {
	return converter.ConvertFile(inputPath, outputPath, format)
}
