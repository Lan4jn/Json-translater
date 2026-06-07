package desktopgui

import "json2table/internal/converter"

func convertFile(inputPath, outputPath, format string) error {
	return converter.ConvertFile(inputPath, outputPath, format)
}
