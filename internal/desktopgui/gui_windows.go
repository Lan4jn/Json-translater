//go:build windows

package desktopgui

import (
	"fmt"
	"path/filepath"
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	cwUseDefault       = 0x80000000
	colorBtnFace       = 15
	idcArrow           = 32512
	swShow             = 5
	wmCommand          = 0x0111
	wmClose            = 0x0010
	wmDestroy          = 0x0002
	wmSetFont          = 0x0030
	bnClicked          = 0
	cbnSelChange       = 1
	wsCaption          = 0x00c00000
	wsSysMenu          = 0x00080000
	wsMinimizeBox      = 0x00020000
	wsVisible          = 0x10000000
	wsChild            = 0x40000000
	wsTabStop          = 0x00010000
	wsExClientEdge     = 0x00000200
	esAutoHScroll      = 0x0080
	bsPushButton       = 0
	bsDefPushButton    = 0x0001
	cbsDropDownList    = 0x0003
	cbsHasStrings      = 0x0200
	ofnOverwritePrompt = 0x00000002
	ofnPathMustExist   = 0x00000800
	ofnFileMustExist   = 0x00001000
	cbAddString        = 0x0143
	cbGetCurSel        = 0x0147
	cbSetCurSel        = 0x014e
	mbOK               = 0
	mbIconInformation  = 0x00000040
	mbIconWarning      = 0x00000030
	mbIconError        = 0x00000010
	fwNormal           = 400
	fwSemiBold         = 600
	defaultCharset     = 1
	clearTypeQuality   = 5
	variablePitch      = 2
)

const (
	inputEditID = 1001 + iota
	outputEditID
	browseInputID
	browseOutputID
	formatComboID
	convertButtonID
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	comdlg32             = syscall.NewLazyDLL("comdlg32.dll")
	gdi32                = syscall.NewLazyDLL("gdi32.dll")
	procRegisterClassEx  = user32.NewProc("RegisterClassExW")
	procCreateWindowEx   = user32.NewProc("CreateWindowExW")
	procDefWindowProc    = user32.NewProc("DefWindowProcW")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procGetMessage       = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage  = user32.NewProc("DispatchMessageW")
	procLoadCursor       = user32.NewProc("LoadCursorW")
	procMessageBox       = user32.NewProc("MessageBoxW")
	procSetWindowText    = user32.NewProc("SetWindowTextW")
	procGetWindowText    = user32.NewProc("GetWindowTextW")
	procGetWindowTextLen = user32.NewProc("GetWindowTextLengthW")
	procSendMessage      = user32.NewProc("SendMessageW")
	procSetFocus         = user32.NewProc("SetFocus")
	procGetModuleHandle  = kernel32.NewProc("GetModuleHandleW")
	procGetOpenFileName  = comdlg32.NewProc("GetOpenFileNameW")
	procGetSaveFileName  = comdlg32.NewProc("GetSaveFileNameW")
	procCreateFont       = gdi32.NewProc("CreateFontW")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type point struct {
	x int32
	y int32
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type openFileName struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

type windowState struct {
	hwnd        uintptr
	inputEdit   uintptr
	outputEdit  uintptr
	formatCombo uintptr
	status      uintptr
	font        uintptr
	titleFont   uintptr
}

var activeWindow *windowState

func Run() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	instance, _, _ := procGetModuleHandle.Call(0)
	className := utf16Ptr("Json2TableNativeWindow")
	cursor, _, _ := procLoadCursor.Call(0, idcArrow)
	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(windowProc),
		hInstance:     instance,
		hCursor:       cursor,
		hbrBackground: colorBtnFace + 1,
		lpszClassName: className,
	}
	if atom, _, err := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); atom == 0 {
		return fmt.Errorf("注册窗口类失败: %w", err)
	}

	hwnd, _, err := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("JSON 转 CSV/Excel"))),
		wsCaption|wsSysMenu|wsMinimizeBox,
		cwUseDefault,
		cwUseDefault,
		760,
		390,
		0,
		0,
		instance,
		0,
	)
	if hwnd == 0 {
		return fmt.Errorf("创建窗口失败: %w", err)
	}

	state := &windowState{hwnd: hwnd, font: createUIFont(), titleFont: createTitleFont()}
	activeWindow = state
	createControls(state, instance)
	procShowWindow.Call(hwnd, swShow)
	procUpdateWindow.Call(hwnd)
	procSetFocus.Call(state.inputEdit)

	var message msg
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
	activeWindow = nil
	return nil
}

func createControls(state *windowState, instance uintptr) {
	title := addControl(state, "STATIC", "JSON 转表格转换器", wsChild|wsVisible, 28, 22, 360, 30, 0, instance)
	if title != 0 && state.titleFont != 0 {
		procSendMessage.Call(title, wmSetFont, state.titleFont, 1)
	}
	addControl(state, "STATIC", "将 JSON 文件导出为 CSV 或 Excel 工作簿", wsChild|wsVisible, 28, 56, 520, 24, 0, instance)

	addControl(state, "STATIC", "JSON 文件", wsChild|wsVisible, 28, 98, 120, 24, 0, instance)
	state.inputEdit = addControlEx(state, wsExClientEdge, "EDIT", "", wsChild|wsVisible|wsTabStop|esAutoHScroll, 28, 122, 560, 30, inputEditID, instance)
	addControl(state, "BUTTON", "选择...", wsChild|wsVisible|wsTabStop|bsPushButton, 606, 121, 112, 32, browseInputID, instance)

	addControl(state, "STATIC", "输出文件", wsChild|wsVisible, 28, 166, 120, 24, 0, instance)
	state.outputEdit = addControlEx(state, wsExClientEdge, "EDIT", "", wsChild|wsVisible|wsTabStop|esAutoHScroll, 28, 190, 560, 30, outputEditID, instance)
	addControl(state, "BUTTON", "另存为...", wsChild|wsVisible|wsTabStop|bsPushButton, 606, 189, 112, 32, browseOutputID, instance)

	addControl(state, "STATIC", "输出格式", wsChild|wsVisible, 28, 234, 120, 24, 0, instance)
	state.formatCombo = addControl(state, "COMBOBOX", "", wsChild|wsVisible|wsTabStop|cbsDropDownList|cbsHasStrings, 28, 258, 240, 160, formatComboID, instance)
	comboAdd(state.formatCombo, "自动：默认 CSV")
	comboAdd(state.formatCombo, "CSV")
	comboAdd(state.formatCombo, "Excel XLSX")
	procSendMessage.Call(state.formatCombo, cbSetCurSel, 0, 0)

	state.status = addControl(state, "STATIC", "准备就绪", wsChild|wsVisible, 28, 302, 540, 24, 0, instance)
	addControl(state, "BUTTON", "开始转换", wsChild|wsVisible|wsTabStop|bsDefPushButton, 606, 292, 112, 36, convertButtonID, instance)
}

func windowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmCommand:
		id := int(wParam & 0xffff)
		notify := int((wParam >> 16) & 0xffff)
		if notify == bnClicked || notify == cbnSelChange {
			handleCommand(id)
		}
		return 0
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		if activeWindow != nil {
			if activeWindow.font != 0 {
				procDeleteObject.Call(activeWindow.font)
				activeWindow.font = 0
			}
			if activeWindow.titleFont != 0 {
				procDeleteObject.Call(activeWindow.titleFont)
				activeWindow.titleFont = 0
			}
		}
		procPostQuitMessage.Call(0)
		return 0
	default:
		ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(message), wParam, lParam)
		return ret
	}
}

func handleCommand(id int) {
	if activeWindow == nil {
		return
	}
	switch id {
	case browseInputID:
		if path, ok := openJSONFile(activeWindow.hwnd); ok {
			setWindowText(activeWindow.inputEdit, path)
			if getWindowText(activeWindow.outputEdit) == "" {
				setWindowText(activeWindow.outputEdit, defaultOutputPathForFormat(path, selectedFormat()))
			}
		}
	case browseOutputID:
		defaultPath := getWindowText(activeWindow.outputEdit)
		if defaultPath == "" {
			defaultPath = defaultOutputPathForFormat(getWindowText(activeWindow.inputEdit), selectedFormat())
		}
		if path, ok := saveOutputFile(activeWindow.hwnd, defaultPath, selectedFormat()); ok {
			setWindowText(activeWindow.outputEdit, path)
		}
	case formatComboID:
		input := getWindowText(activeWindow.inputEdit)
		if input != "" {
			setWindowText(activeWindow.outputEdit, defaultOutputPathForFormat(input, selectedFormat()))
		}
	case convertButtonID:
		convertFromWindow()
	}
}

func convertFromWindow() {
	inputPath := getWindowText(activeWindow.inputEdit)
	if inputPath == "" {
		messageBox(activeWindow.hwnd, "请选择 JSON 文件。", "提示", mbOK|mbIconWarning)
		return
	}
	outputPath := getWindowText(activeWindow.outputEdit)
	format := selectedFormat()
	if outputPath == "" {
		outputPath = defaultOutputPathForFormat(inputPath, format)
		setWindowText(activeWindow.outputEdit, outputPath)
	}
	if err := convertFile(inputPath, outputPath, format); err != nil {
		setWindowText(activeWindow.status, "转换失败")
		messageBox(activeWindow.hwnd, "转换失败: "+err.Error(), "错误", mbOK|mbIconError)
		return
	}
	setWindowText(activeWindow.status, "已保存到 "+outputPath)
	messageBox(activeWindow.hwnd, "已保存到\n"+outputPath, "转换完成", mbOK|mbIconInformation)
}

func selectedFormat() string {
	if activeWindow == nil || activeWindow.formatCombo == 0 {
		return ""
	}
	index, _, _ := procSendMessage.Call(activeWindow.formatCombo, cbGetCurSel, 0, 0)
	switch index {
	case 1:
		return "csv"
	case 2:
		return "xlsx"
	default:
		return ""
	}
}

func addControl(state *windowState, className, text string, style uintptr, x, y, width, height int32, id int, instance uintptr) uintptr {
	return addControlEx(state, 0, className, text, style, x, y, width, height, id, instance)
}

func addControlEx(state *windowState, exStyle uintptr, className, text string, style uintptr, x, y, width, height int32, id int, instance uintptr) uintptr {
	hwnd, _, _ := procCreateWindowEx.Call(
		exStyle,
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		style,
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		state.hwnd,
		uintptr(id),
		instance,
		0,
	)
	if hwnd != 0 && state.font != 0 {
		procSendMessage.Call(hwnd, wmSetFont, state.font, 1)
	}
	return hwnd
}

func comboAdd(hwnd uintptr, text string) {
	procSendMessage.Call(hwnd, cbAddString, 0, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func setWindowText(hwnd uintptr, text string) {
	procSetWindowText.Call(hwnd, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getWindowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLen.Call(hwnd)
	buffer := make([]uint16, length+1)
	procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buffer[0])), length+1)
	return syscall.UTF16ToString(buffer)
}

func openJSONFile(owner uintptr) (string, bool) {
	return fileDialog(owner, true, "", "JSON 文件 (*.json)\x00*.json\x00所有文件 (*.*)\x00*.*\x00\x00", "json", 1)
}

func saveOutputFile(owner uintptr, defaultPath, format string) (string, bool) {
	defExt := "csv"
	filterIndex := uint32(1)
	if format == "xlsx" || filepath.Ext(defaultPath) == ".xlsx" {
		defExt = "xlsx"
		filterIndex = 2
	}
	return fileDialog(owner, false, defaultPath, "CSV 文件 (*.csv)\x00*.csv\x00Excel 工作簿 (*.xlsx)\x00*.xlsx\x00所有文件 (*.*)\x00*.*\x00\x00", defExt, filterIndex)
}

func fileDialog(owner uintptr, open bool, initialPath, filter, defExt string, filterIndex uint32) (string, bool) {
	buffer := make([]uint16, 4096)
	if initialPath != "" {
		copy(buffer, utf16String(initialPath))
	}
	filterUTF16 := utf16String(filter)
	defExtUTF16 := utf16String(defExt)
	ofn := openFileName{
		lStructSize:  uint32(unsafe.Sizeof(openFileName{})),
		hwndOwner:    owner,
		lpstrFilter:  &filterUTF16[0],
		nFilterIndex: filterIndex,
		lpstrFile:    &buffer[0],
		nMaxFile:     uint32(len(buffer)),
		lpstrDefExt:  &defExtUTF16[0],
	}
	if open {
		ofn.flags = ofnPathMustExist | ofnFileMustExist
		ret, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
		return syscall.UTF16ToString(buffer), ret != 0
	}
	ofn.flags = ofnOverwritePrompt
	ret, _, _ := procGetSaveFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	return syscall.UTF16ToString(buffer), ret != 0
}

func createUIFont() uintptr {
	return createFont(-15, fwNormal)
}

func createTitleFont() uintptr {
	return createFont(-20, fwSemiBold)
}

func createFont(height int32, weight uintptr) uintptr {
	font, _, _ := procCreateFont.Call(
		signedInt32Param(height),
		0,
		0,
		0,
		weight,
		0,
		0,
		0,
		defaultCharset,
		0,
		0,
		clearTypeQuality,
		variablePitch,
		uintptr(unsafe.Pointer(utf16Ptr("Microsoft YaHei UI"))),
	)
	return font
}

func signedInt32Param(value int32) uintptr {
	return uintptr(int64(value))
}

func messageBox(owner uintptr, text, title string, flags uintptr) {
	procMessageBox.Call(owner, uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(unsafe.Pointer(utf16Ptr(title))), flags)
}

func utf16Ptr(value string) *uint16 {
	encoded := utf16String(value)
	return &encoded[0]
}

func utf16String(value string) []uint16 {
	return append(utf16.Encode([]rune(value)), 0)
}
