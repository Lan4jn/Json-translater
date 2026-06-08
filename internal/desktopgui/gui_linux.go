//go:build linux

package desktopgui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

var errCanceled = errors.New("canceled")

var errLinuxGUIUnavailable = errors.New("Linux 图形界面模式需要安装 zenity、kdialog，或启用 xdg-desktop-portal")

const (
	portalBusName              = "org.freedesktop.portal.Desktop"
	portalObjectPath           = dbus.ObjectPath("/org/freedesktop/portal/desktop")
	portalFileChooserInterface = "org.freedesktop.portal.FileChooser"
	portalRequestResponse      = "org.freedesktop.portal.Request.Response"
)

func Run() error {
	inputPath, err := linuxOpenFile()
	if errors.Is(err, errCanceled) {
		return nil
	}
	if err != nil {
		if !errors.Is(err, errLinuxGUIUnavailable) {
			_ = showError(err.Error())
		}
		return err
	}

	outputPath, err := linuxSaveFile(defaultOutputPath(inputPath))
	if errors.Is(err, errCanceled) {
		return nil
	}
	if err != nil {
		if !errors.Is(err, errLinuxGUIUnavailable) {
			_ = showError(err.Error())
		}
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
	if shouldUsePortal() {
		path, err := portalOpenFile("选择 JSON 文件")
		return path, err
	}
	return "", errLinuxGUIUnavailable
}

func linuxSaveFile(defaultPath string) (string, error) {
	if commandExists("zenity") {
		return runPathCommand("zenity", "--file-selection", "--save", "--confirm-overwrite", "--title", "保存转换结果", "--filename", defaultPath)
	}
	if commandExists("kdialog") {
		return runPathCommand("kdialog", "--title", "保存转换结果", "--getsavefilename", defaultPath, "*.csv *.xlsx")
	}
	if shouldUsePortal() {
		path, err := portalSaveFile("保存转换结果", defaultPath)
		return path, err
	}
	return "", errLinuxGUIUnavailable
}

func showInfo(title, message string) error {
	if commandExists("zenity") {
		return exec.Command("zenity", "--info", "--title", title, "--text", message).Run()
	}
	if commandExists("kdialog") {
		return exec.Command("kdialog", "--title", title, "--msgbox", message).Run()
	}
	if commandExists("notify-send") {
		return exec.Command("notify-send", title, message).Run()
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
	if commandExists("notify-send") {
		return exec.Command("notify-send", "JSON 转换错误", message).Run()
	}
	return nil
}

func portalOpenFile(title string) (string, error) {
	return portalFileChooser("OpenFile", title, map[string]dbus.Variant{
		"accept_label": dbus.MakeVariant("打开"),
	})
}

func portalSaveFile(title, defaultPath string) (string, error) {
	options := map[string]dbus.Variant{
		"accept_label": dbus.MakeVariant("保存"),
	}
	if defaultPath != "" {
		options["current_name"] = dbus.MakeVariant(filepath.Base(defaultPath))
		if dir := filepath.Dir(defaultPath); dir != "." && dir != string(filepath.Separator) {
			options["current_folder"] = dbus.MakeVariant(append([]byte(dir), 0))
		}
	}
	return portalFileChooser("SaveFile", title, options)
}

func portalFileChooser(method, title string, options map[string]dbus.Variant) (string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return "", fmt.Errorf("connect session bus: %w", err)
	}
	defer conn.Close()

	token := fmt.Sprintf("json2table_%d_%d", os.Getpid(), time.Now().UnixNano())
	options["handle_token"] = dbus.MakeVariant(token)

	signals := make(chan *dbus.Signal, 8)
	conn.Signal(signals)
	defer conn.RemoveSignal(signals)

	match := "type='signal',interface='org.freedesktop.portal.Request',member='Response'"
	if call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, match); call.Err != nil {
		return "", fmt.Errorf("listen portal response: %w", call.Err)
	}
	defer conn.BusObject().Call("org.freedesktop.DBus.RemoveMatch", 0, match)

	var requestPath dbus.ObjectPath
	call := conn.Object(portalBusName, portalObjectPath).Call(
		portalFileChooserInterface+"."+method,
		0,
		"",
		title,
		options,
	)
	if call.Err != nil {
		return "", fmt.Errorf("open portal file chooser: %w", call.Err)
	}
	if err := call.Store(&requestPath); err != nil {
		return "", fmt.Errorf("read portal request handle: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("wait portal response: %w", ctx.Err())
		case signal := <-signals:
			path, done, err := portalSelectedPathFromSignal(signal, requestPath)
			if !done {
				continue
			}
			return path, err
		}
	}
}

func portalSelectedPathFromSignal(signal *dbus.Signal, requestPath dbus.ObjectPath) (string, bool, error) {
	if signal == nil || signal.Name != portalRequestResponse || signal.Path != requestPath {
		return "", false, nil
	}
	if len(signal.Body) < 2 {
		return "", true, errors.New("portal response is incomplete")
	}

	response, ok := signal.Body[0].(uint32)
	if !ok {
		return "", true, fmt.Errorf("unexpected portal response code type %T", signal.Body[0])
	}
	if response != 0 {
		return "", true, errCanceled
	}

	results, ok := signal.Body[1].(map[string]dbus.Variant)
	if !ok {
		return "", true, fmt.Errorf("unexpected portal response result type %T", signal.Body[1])
	}
	urisVariant, ok := results["uris"]
	if !ok {
		return "", true, errors.New("portal response did not include file URI")
	}
	uris, ok := urisVariant.Value().([]string)
	if !ok {
		return "", true, fmt.Errorf("unexpected portal URI type %T", urisVariant.Value())
	}
	path, err := selectedPathFromPortalURIs(uris)
	return path, true, err
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

func shouldUsePortal() bool {
	return os.Getenv("JSON2TABLE_USE_PORTAL") == "1"
}
