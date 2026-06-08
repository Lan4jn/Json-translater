//go:build linux

package desktopgui

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var errCanceled = errors.New("canceled")

func Run() error {
	inputPath, err := linuxOpenFile()
	if errors.Is(err, errCanceled) {
		return nil
	}
	if err != nil {
		_ = showError(err.Error())
		return err
	}

	outputPath, err := linuxSaveFile(defaultOutputPath(inputPath))
	if errors.Is(err, errCanceled) {
		return nil
	}
	if err != nil {
		_ = showError(err.Error())
		return err
	}

	if err := convertFile(inputPath, outputPath, ""); err != nil {
		_ = showError("转换失败: " + err.Error())
		return err
	}
	return showInfo("转换完成", fmt.Sprintf("已保存到\n%s", outputPath))
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

func showInfo(title, message string) error {
	if commandExists("zenity") {
		return exec.Command("zenity", "--info", "--title", title, "--text", message).Run()
	}
	if commandExists("kdialog") {
		return exec.Command("kdialog", "--title", title, "--msgbox", message).Run()
	}
	return nil
}

func showError(message string) error {
	if commandExists("zenity") {
		return exec.Command("zenity", "--error", "--title", "错误", "--text", message).Run()
	}
	if commandExists("kdialog") {
		return exec.Command("kdialog", "--title", "错误", "--error", message).Run()
	}
	return nil
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
	path := strings.TrimSpace(string(output))
	if path == "" {
		return "", errCanceled
	}
	return path, nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
