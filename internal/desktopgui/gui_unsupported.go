//go:build !windows && !linux

package desktopgui

import (
	"fmt"
	"runtime"
)

func Run() error {
	return fmt.Errorf("当前系统 %s 不支持图形界面模式，请使用命令行参数", runtime.GOOS)
}
