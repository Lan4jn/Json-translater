package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunConvertsJSONToCSV(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.json")
	output := filepath.Join(dir, "output.csv")
	if err := os.WriteFile(input, []byte(`[{"name":"Ada"}]`), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"-input", input, "-output", output}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, stderr.String())
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(data) != "name\nAda\n" {
		t.Fatalf("unexpected output: %q", string(data))
	}
	if !strings.Contains(stdout.String(), "converted") {
		t.Fatalf("expected success message, got %q", stdout.String())
	}
}

func TestRunRejectsMissingRequiredFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"-input", "input.json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "output path is required") {
		t.Fatalf("expected missing output error, got %q", stderr.String())
	}
}
