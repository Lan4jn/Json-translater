package webui

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConvertDataWritesCSV(t *testing.T) {
	var output bytes.Buffer
	err := convertData([]byte(`[{"name":"Ada","age":37}]`), &output, "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := output.String()
	if !strings.Contains(got, "age,name") || !strings.Contains(got, "37,Ada") {
		t.Fatalf("unexpected csv output: %q", got)
	}
}

func TestOutputNameUsesSelectedFormat(t *testing.T) {
	if got := outputName("people.json", "xlsx"); got != "people.xlsx" {
		t.Fatalf("expected people.xlsx, got %q", got)
	}
	if got := outputName("people.json", "csv"); got != "people.csv" {
		t.Fatalf("expected people.csv, got %q", got)
	}
}

func TestIndexReturnsUploadPage(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "JSON 转 CSV/Excel") {
		t.Fatalf("expected upload page, got %q", response.Body.String())
	}
}

func TestConvertEndpointReturnsDownload(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("json_file", "people.json")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = file.Write([]byte(`[{"name":"Ada","age":37}]`))
	_ = writer.WriteField("format", "csv")
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/convert", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()

	routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if disposition := response.Header().Get("Content-Disposition"); !strings.Contains(disposition, "people.csv") {
		t.Fatalf("expected people.csv disposition, got %q", disposition)
	}
	if !strings.Contains(response.Body.String(), "37,Ada") {
		t.Fatalf("unexpected csv body: %q", response.Body.String())
	}
}

func TestWaitForReadyChecksHTTPServer(t *testing.T) {
	server := httptest.NewServer(routes())
	defer server.Close()

	serveErr := make(chan error, 1)
	if err := waitForReady(server.URL, serveErr, time.Second); err != nil {
		t.Fatalf("expected server to be ready: %v", err)
	}
}

func TestBrowserCommandsForLinuxIncludeCommonOpeners(t *testing.T) {
	commands := browserCommands("linux", "http://127.0.0.1:1")
	names := make(map[string]bool, len(commands))
	for _, command := range commands {
		names[command.name] = true
	}
	for _, name := range []string{"xdg-open", "gio", "deepin-browser", "firefox"} {
		if !names[name] {
			t.Fatalf("expected linux browser command %q in %#v", name, commands)
		}
	}
}
