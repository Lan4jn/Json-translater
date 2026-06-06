package desktopgui

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"json2table/internal/converter"
)

var errCanceled = errors.New("canceled")

func Run() error {
	inputPath, err := pickInputFile()
	if errors.Is(err, errCanceled) {
		return nil
	}
	if err != nil {
		_ = showError(err.Error())
		return err
	}

	outputPath, err := pickOutputFile(defaultOutputPath(inputPath))
	if errors.Is(err, errCanceled) {
		return nil
	}
	if err != nil {
		_ = showError(err.Error())
		return err
	}

	if err := converter.ConvertFile(inputPath, outputPath, ""); err != nil {
		_ = showError("转换失败: " + err.Error())
		return err
	}

	return showInfo("转换完成", "已保存到\n"+outputPath)
}

func defaultOutputPath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if base == "" || base == "." {
		base = "output"
	}
	return filepath.Join(dir, base+".csv")
}

func pickInputFile() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return runPowerShellPath(`
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Title = '选择 JSON 文件'
$dialog.Filter = 'JSON 文件 (*.json)|*.json|所有文件 (*.*)|*.*'
$dialog.Multiselect = $false
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
  [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
  Write-Output $dialog.FileName
  exit 0
}
exit 2
`)
	case "linux":
		return linuxOpenFile()
	default:
		return "", fmt.Errorf("当前系统 %s 不支持图形界面模式，请使用命令行参数", runtime.GOOS)
	}
}

func pickOutputFile(defaultPath string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.SaveFileDialog
$dialog.Title = '保存转换结果'
$dialog.Filter = 'CSV 文件 (*.csv)|*.csv|Excel 工作簿 (*.xlsx)|*.xlsx'
$dialog.AddExtension = $true
$dialog.OverwritePrompt = $true
$dialog.FileName = %s
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
  [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
  Write-Output $dialog.FileName
  exit 0
}
exit 2
`, powerShellString(defaultPath))
		return runPowerShellPath(script)
	case "linux":
		return linuxSaveFile(defaultPath)
	default:
		return "", fmt.Errorf("当前系统 %s 不支持图形界面模式，请使用命令行参数", runtime.GOOS)
	}
}

func showInfo(title, message string) error {
	switch runtime.GOOS {
	case "windows":
		_, err := runPowerShellRaw(fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
[System.Windows.Forms.MessageBox]::Show(%s, %s, [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information) | Out-Null
`, powerShellString(message), powerShellString(title)))
		return err
	case "linux":
		if commandExists("zenity") {
			return exec.Command("zenity", "--info", "--title", title, "--text", message).Run()
		}
		if commandExists("kdialog") {
			return exec.Command("kdialog", "--title", title, "--msgbox", message).Run()
		}
	}
	return nil
}

func showError(message string) error {
	switch runtime.GOOS {
	case "windows":
		_, err := runPowerShellRaw(fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
[System.Windows.Forms.MessageBox]::Show(%s, '错误', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Error) | Out-Null
`, powerShellString(message)))
		return err
	case "linux":
		if commandExists("zenity") {
			return exec.Command("zenity", "--error", "--title", "错误", "--text", message).Run()
		}
		if commandExists("kdialog") {
			return exec.Command("kdialog", "--title", "错误", "--error", message).Run()
		}
	}
	return nil
}

func linuxOpenFile() (string, error) {
	if commandExists("zenity") {
		return runPathCommand("zenity", "--file-selection", "--title", "选择 JSON 文件", "--file-filter", "JSON files | *.json")
	}
	if commandExists("kdialog") {
		return runPathCommand("kdialog", "--title", "选择 JSON 文件", "--getopenfilename", ".", "*.json")
	}
	return "", errors.New("Linux 图形界面模式需要安装 zenity 或 kdialog")
}

func linuxSaveFile(defaultPath string) (string, error) {
	if commandExists("zenity") {
		return runPathCommand("zenity", "--file-selection", "--save", "--confirm-overwrite", "--title", "保存转换结果", "--filename", defaultPath)
	}
	if commandExists("kdialog") {
		return runPathCommand("kdialog", "--title", "保存转换结果", "--getsavefilename", defaultPath, "*.csv *.xlsx")
	}
	return "", errors.New("Linux 图形界面模式需要安装 zenity 或 kdialog")
}

func runPowerShellPath(script string) (string, error) {
	output, err := runPowerShellRaw(script)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 2 {
			return "", errCanceled
		}
		return "", err
	}
	return cleanPath(output)
}

func runPowerShellRaw(script string) (string, error) {
	output, err := exec.Command("powershell", "-NoProfile", "-STA", "-Command", script).Output()
	return string(output), err
}

func runPathCommand(name string, args ...string) (string, error) {
	output, err := exec.Command(name, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", errCanceled
		}
		return "", err
	}
	return cleanPath(string(output))
}

func cleanPath(output string) (string, error) {
	path := strings.TrimSpace(output)
	if path == "" {
		return "", errCanceled
	}
	return path, nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func powerShellString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
