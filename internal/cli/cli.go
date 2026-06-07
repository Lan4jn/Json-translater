package cli

import (
	"flag"
	"fmt"
	"io"

	"json2table/internal/converter"
)

func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("json2table", flag.ContinueOnError)
	flags.SetOutput(stderr)

	inputPath := flags.String("input", "", "input JSON file path")
	outputPath := flags.String("output", "", "output CSV or XLSX file path")
	format := flags.String("format", "", "output format: csv or xlsx")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *inputPath == "" {
		fmt.Fprintln(stderr, "input path is required")
		return 2
	}
	if *outputPath == "" {
		fmt.Fprintln(stderr, "output path is required")
		return 2
	}

	if err := converter.ConvertFile(*inputPath, *outputPath, *format); err != nil {
		fmt.Fprintf(stderr, "conversion failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "converted %s to %s\n", *inputPath, *outputPath)
	return 0
}
