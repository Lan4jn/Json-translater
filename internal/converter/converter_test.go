package converter

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveFormatDefaultsToCSV(t *testing.T) {
	format, err := ResolveFormat("out", "")
	if err != nil {
		t.Fatalf("ResolveFormat returned error: %v", err)
	}
	if format != CSV {
		t.Fatalf("expected CSV, got %q", format)
	}
}

func TestResolveFormatUsesOutputExtensionAndExplicitOverride(t *testing.T) {
	format, err := ResolveFormat("out.xlsx", "")
	if err != nil {
		t.Fatalf("ResolveFormat returned error: %v", err)
	}
	if format != XLSX {
		t.Fatalf("expected XLSX from extension, got %q", format)
	}

	format, err = ResolveFormat("out.xlsx", "csv")
	if err != nil {
		t.Fatalf("ResolveFormat returned error: %v", err)
	}
	if format != CSV {
		t.Fatalf("expected explicit CSV override, got %q", format)
	}
}

func TestWriteCSVFromJSONArrayUsesSortedUnionColumns(t *testing.T) {
	table, err := DecodeJSON([]byte(`[
		{"name":"Ada","age":37,"tags":["math","code"]},
		{"name":"Linus","active":true,"meta":{"country":"FI"}}
	]`))
	if err != nil {
		t.Fatalf("DecodeJSON returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, table); err != nil {
		t.Fatalf("WriteCSV returned error: %v", err)
	}

	rows, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
	if err != nil {
		t.Fatalf("CSV output is invalid: %v\n%s", err, buf.String())
	}

	want := [][]string{
		{"active", "age", "meta", "name", "tags"},
		{"", "37", "", "Ada", `["math","code"]`},
		{"true", "", `{"country":"FI"}`, "Linus", ""},
	}
	if len(rows) != len(want) {
		t.Fatalf("expected %d rows, got %d: %#v", len(want), len(rows), rows)
	}
	for i := range want {
		for j := range want[i] {
			if rows[i][j] != want[i][j] {
				t.Fatalf("row %d col %d: expected %q, got %q", i, j, want[i][j], rows[i][j])
			}
		}
	}
}

func TestDecodeJSONAcceptsSingleObject(t *testing.T) {
	table, err := DecodeJSON([]byte(`{"name":"Grace","age":85}`))
	if err != nil {
		t.Fatalf("DecodeJSON returned error: %v", err)
	}
	if len(table.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(table.Rows))
	}
	if got := table.Rows[0]["name"]; got != "Grace" {
		t.Fatalf("expected name Grace, got %q", got)
	}
}

func TestDecodeJSONRejectsUnsupportedTopLevelValue(t *testing.T) {
	_, err := DecodeJSON([]byte(`"not a table"`))
	if err == nil {
		t.Fatal("expected error for unsupported top-level JSON value")
	}
}

func TestConvertFileDefaultsToCSV(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.json")
	output := filepath.Join(dir, "output")
	if err := os.WriteFile(input, []byte(`[{"b":2,"a":1}]`), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := ConvertFile(input, output, ""); err != nil {
		t.Fatalf("ConvertFile returned error: %v", err)
	}

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(got) != "a,b\n1,2\n" {
		t.Fatalf("unexpected CSV output: %q", string(got))
	}
}

func TestConvertFileWritesXLSX(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.json")
	output := filepath.Join(dir, "output.xlsx")
	if err := os.WriteFile(input, []byte(`[{"name":"Ada","age":37}]`), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := ConvertFile(input, output, ""); err != nil {
		t.Fatalf("ConvertFile returned error: %v", err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("output is not a valid xlsx zip: %v", err)
	}

	foundSheet := false
	for _, file := range reader.File {
		if file.Name == "xl/worksheets/sheet1.xml" {
			foundSheet = true
			break
		}
	}
	if !foundSheet {
		t.Fatal("xlsx file does not contain sheet1.xml")
	}
}
