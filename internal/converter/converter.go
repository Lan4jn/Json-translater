package converter

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Format string

const (
	CSV  Format = "csv"
	XLSX Format = "xlsx"
)

type Table struct {
	Columns []string
	Rows    []map[string]string
}

func ResolveFormat(outputPath, explicit string) (Format, error) {
	if explicit != "" {
		return parseFormat(explicit)
	}
	if strings.EqualFold(filepath.Ext(outputPath), ".xlsx") {
		return XLSX, nil
	}
	return CSV, nil
}

func ConvertFile(inputPath, outputPath, explicitFormat string) error {
	if inputPath == "" {
		return errors.New("input path is required")
	}
	if outputPath == "" {
		return errors.New("output path is required")
	}

	format, err := ResolveFormat(outputPath, explicitFormat)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input file: %w", err)
	}
	table, err := DecodeJSON(data)
	if err != nil {
		return err
	}
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer output.Close()

	switch format {
	case CSV:
		return WriteCSV(output, table)
	case XLSX:
		return WriteXLSX(output, table)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func DecodeJSON(data []byte) (Table, error) {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()

	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return Table{}, fmt.Errorf("parse json: %w", err)
	}

	var objects []map[string]any
	switch value := raw.(type) {
	case map[string]any:
		objects = []map[string]any{value}
	case []any:
		for i, item := range value {
			object, ok := item.(map[string]any)
			if !ok {
				return Table{}, fmt.Errorf("array item %d is %T, expected object", i, item)
			}
			objects = append(objects, object)
		}
	default:
		return Table{}, fmt.Errorf("top-level JSON value is %T, expected object or array of objects", raw)
	}

	columns := collectColumns(objects)
	rows := make([]map[string]string, 0, len(objects))
	for _, object := range objects {
		row := make(map[string]string, len(object))
		for key, value := range object {
			row[key] = cellString(value)
		}
		rows = append(rows, row)
	}
	return Table{Columns: columns, Rows: rows}, nil
}

func WriteCSV(w io.Writer, table Table) error {
	writer := csv.NewWriter(w)
	if err := writer.Write(table.Columns); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	for _, row := range table.Rows {
		record := make([]string, len(table.Columns))
		for i, column := range table.Columns {
			record[i] = row[column]
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}
	return nil
}

func WriteXLSX(w io.Writer, table Table) error {
	file := excelize.NewFile()
	defer file.Close()

	sheet := file.GetSheetName(0)
	for columnIndex, column := range table.Columns {
		cell, err := excelize.CoordinatesToCellName(columnIndex+1, 1)
		if err != nil {
			return fmt.Errorf("xlsx header cell: %w", err)
		}
		if err := file.SetCellValue(sheet, cell, column); err != nil {
			return fmt.Errorf("write xlsx header: %w", err)
		}
	}
	for rowIndex, row := range table.Rows {
		for columnIndex, column := range table.Columns {
			cell, err := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+2)
			if err != nil {
				return fmt.Errorf("xlsx data cell: %w", err)
			}
			if err := file.SetCellValue(sheet, cell, row[column]); err != nil {
				return fmt.Errorf("write xlsx row: %w", err)
			}
		}
	}
	if err := file.Write(w); err != nil {
		return fmt.Errorf("write xlsx file: %w", err)
	}
	return nil
}

func parseFormat(value string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "csv":
		return CSV, nil
	case "xlsx", "excel":
		return XLSX, nil
	default:
		return "", fmt.Errorf("unsupported output format %q, use csv or xlsx", value)
	}
}

func collectColumns(objects []map[string]any) []string {
	seen := make(map[string]bool)
	for _, object := range objects {
		for key := range object {
			seen[key] = true
		}
	}
	columns := make([]string, 0, len(seen))
	for key := range seen {
		columns = append(columns, key)
	}
	sort.Strings(columns)
	return columns
}

func cellString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case bool:
		return strconv.FormatBool(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(data)
	}
}
