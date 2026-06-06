# json2table

`json2table` 是一个将 JSON 文件转换为 CSV 或 Excel XLSX 的 Go 可执行程序。

## 使用方式

不带参数运行时会启动图形界面。Windows 使用系统文件选择/保存对话框；Linux 使用 `zenity` 或 `kdialog`：

```powershell
.\json2table.exe
```

带参数运行时使用命令行模式：

```powershell
.\json2table.exe -input data.json -output out.csv
.\json2table.exe -input data.json -output out.xlsx
.\json2table.exe -input data.json -output out -format csv
```

`-format` 可选，支持 `csv`、`xlsx`。不传时默认 CSV；输出文件名为 `.xlsx` 时自动生成 Excel。

Linux 图形界面模式需要系统已安装 `zenity` 或 `kdialog`。如果没有这些组件，仍可使用命令行模式。

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
