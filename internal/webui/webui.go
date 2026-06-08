package webui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"json2table/internal/converter"
)

const maxUploadSize = 32 << 20

func Run() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("启动 WebUI 端口失败: %w", err)
	}
	defer listener.Close()

	server := &http.Server{
		Handler:           routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	defer server.Close()

	serveErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	url := "http://" + listener.Addr().String()
	if err := waitForReady(url, serveErr, 5*time.Second); err != nil {
		return err
	}
	if err := openBrowser(url); err != nil {
		fmt.Printf("WebUI 已启动，请在浏览器打开：%s\n", url)
		_ = showOpenHint("WebUI 已启动", "请在浏览器打开: "+url)
	} else {
		fmt.Printf("WebUI 已启动：%s\n", url)
		_ = showOpenHint("WebUI 已启动", "浏览器地址: "+url)
	}

	return <-serveErr
}

func waitForReady(url string, serveErr <-chan error, timeout time.Duration) error {
	deadline := time.After(timeout)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for {
		select {
		case err := <-serveErr:
			if err == nil {
				return errors.New("WebUI 服务启动后立即退出")
			}
			return fmt.Errorf("WebUI 服务启动失败: %w", err)
		case <-deadline:
			return errors.New("WebUI 服务启动超时，端口未响应")
		default:
			response, err := client.Get(url)
			if err == nil {
				_ = response.Body.Close()
				if response.StatusCode == http.StatusOK {
					return nil
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/convert", handleConvert)
	return mux
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pageTemplate.Execute(w, nil)
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "读取上传文件失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("json_file")
	if err != nil {
		http.Error(w, "请选择 JSON 文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "读取上传文件失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	format := r.FormValue("format")
	outputName := outputName(header.Filename, format)
	var output bytes.Buffer
	if err := convertData(data, &output, format); err != nil {
		http.Error(w, "转换失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, outputName))
	if format == "xlsx" {
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	} else {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	}
	_, _ = w.Write(output.Bytes())
}

func convertData(data []byte, output io.Writer, format string) error {
	table, err := converter.DecodeJSON(data)
	if err != nil {
		return err
	}
	switch format {
	case "", "csv":
		return converter.WriteCSV(output, table)
	case "xlsx":
		return converter.WriteXLSX(output, table)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func outputName(inputName, format string) string {
	base := inputName
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] == '.' {
			base = base[:i]
			break
		}
	}
	if base == "" {
		base = "output"
	}
	if format == "xlsx" {
		return base + ".xlsx"
	}
	return base + ".csv"
}

func openBrowser(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastErr error
	for _, candidate := range browserCommands(runtime.GOOS, url) {
		if _, err := exec.LookPath(candidate.name); err != nil {
			lastErr = err
			continue
		}
		if err := exec.CommandContext(ctx, candidate.name, candidate.args...).Run(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("unsupported browser platform %s", runtime.GOOS)
}

type browserCommand struct {
	name string
	args []string
}

func browserCommands(goos, url string) []browserCommand {
	switch goos {
	case "windows":
		return []browserCommand{{name: "rundll32", args: []string{"url.dll,FileProtocolHandler", url}}}
	case "linux":
		commands := []browserCommand{}
		if browser := os.Getenv("BROWSER"); browser != "" {
			commands = append(commands, browserCommand{name: browser, args: []string{url}})
		}
		return append(commands,
			browserCommand{name: "xdg-open", args: []string{url}},
			browserCommand{name: "gio", args: []string{"open", url}},
			browserCommand{name: "sensible-browser", args: []string{url}},
			browserCommand{name: "x-www-browser", args: []string{url}},
			browserCommand{name: "deepin-browser", args: []string{url}},
			browserCommand{name: "google-chrome", args: []string{url}},
			browserCommand{name: "chromium", args: []string{url}},
			browserCommand{name: "firefox", args: []string{url}},
		)
	default:
		return nil
	}
}

func showOpenHint(title, message string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	if _, err := exec.LookPath("notify-send"); err == nil {
		return exec.Command("notify-send", title, message).Start()
	}
	return nil
}

var pageTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>JSON 转 CSV/Excel</title>
  <style>
    :root { color-scheme: light; font-family: "Segoe UI", "Microsoft YaHei UI", sans-serif; }
    body { margin: 0; background: #f4f6f8; color: #17202a; }
    main { max-width: 760px; margin: 48px auto; padding: 0 24px; }
    h1 { margin: 0 0 8px; font-size: 28px; font-weight: 700; }
    p { margin: 0 0 28px; color: #52606d; }
    form { background: #fff; border: 1px solid #d8dee6; border-radius: 8px; padding: 28px; box-shadow: 0 12px 30px rgba(23,32,42,.08); }
    label { display: block; margin-bottom: 8px; font-weight: 600; }
    input[type=file], select { width: 100%; box-sizing: border-box; margin-bottom: 22px; padding: 10px; border: 1px solid #b8c2cc; border-radius: 6px; font: inherit; background: white; }
    button { min-width: 132px; height: 40px; border: 0; border-radius: 6px; background: #1769aa; color: white; font: inherit; font-weight: 600; cursor: pointer; }
    button:hover { background: #125589; }
  </style>
</head>
<body>
  <main>
    <h1>JSON 转 CSV/Excel</h1>
    <p>选择本地 JSON 文件，转换结果会由浏览器下载。</p>
    <form method="post" action="/convert" enctype="multipart/form-data">
      <label for="json_file">JSON 文件</label>
      <input id="json_file" name="json_file" type="file" accept=".json,application/json" required>
      <label for="format">输出格式</label>
      <select id="format" name="format">
        <option value="csv">CSV</option>
        <option value="xlsx">Excel XLSX</option>
      </select>
      <button type="submit">开始转换</button>
    </form>
  </main>
</body>
</html>`))
