# json2table

`json2table` 是一个将 JSON 文件转换为 CSV 或 Excel XLSX 的 Go 可执行程序。

## 使用方式

不带参数运行时会启动图形界面。Windows 使用程序内置的原生桌面窗口；Linux 图形环境会优先使用 `zenity` 或 `kdialog`。如果桌面 GUI 组件不可用，程序会自动切换到本地 WebUI：

```powershell
.\json2table.exe
```

也可以直接启动浏览器版 WebUI：

```powershell
.\json2table.exe --webui
```

带参数运行时使用命令行模式：

```powershell
.\json2table.exe -input data.json -output out.csv
.\json2table.exe -input data.json -output out.xlsx
.\json2table.exe -input data.json -output out -format csv
```

`-format` 可选，支持 `csv`、`xlsx`。不传时默认 CSV；输出文件名为 `.xlsx` 时自动生成 Excel。

Linux 原生图形界面模式建议系统已安装 `zenity` 或 `kdialog`。统信 UOS / Deepin 等环境如果没有这些组件，会回退到本地 WebUI 并自动打开浏览器。

常见安装方式：

```bash
# Debian / Ubuntu
sudo apt install zenity

# Fedora
sudo dnf install zenity

# Arch
sudo pacman -S zenity
```

如需强制尝试 XDG Desktop Portal 文件选择器，可设置环境变量：

```bash
JSON2TABLE_USE_PORTAL=1 ./json2table-linux-amd64
```

如果没有这些组件，仍可使用命令行模式。

## JSON 结构

支持顶层对象或对象数组：

```json
{"name":"Ada","age":37}
```

```json
[
  {"name":"Ada","age":37},
  {"name":"Linus","active":true}
]
```

列名会取所有对象字段的并集，并按字段名排序。嵌套对象和数组会以紧凑 JSON 字符串写入单元格。

## 构建

Windows:

```powershell
go build -o dist/json2table-windows-amd64.exe .
```

Linux:

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o dist/json2table-linux-amd64 .
```

在 Linux 本机也可以直接运行：

```bash
go build -o dist/json2table-linux-amd64 .
```
