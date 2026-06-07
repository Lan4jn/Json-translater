//go:build windows

package desktopgui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}

	script := windowsFormScript(exePath)
	scriptPath, err := writeTempPowerShellScript(script)
	if err != nil {
		return fmt.Errorf("创建桌面界面脚本失败: %w", err)
	}
	defer os.Remove(scriptPath)

	name, args := windowsPowerShellCommand(scriptPath)
	cmd := exec.Command(name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("启动桌面界面失败: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func writeTempPowerShellScript(script string) (string, error) {
	file, err := os.CreateTemp("", "json2table-gui-*.ps1")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(script); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

func windowsPowerShellCommand(scriptPath string) (string, []string) {
	if pwsh, err := exec.LookPath("pwsh.exe"); err == nil {
		return pwsh, []string{"-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
	}

	home := os.Getenv("USERPROFILE")
	if home != "" {
		pwshBat := filepath.Join(home, "bin", "pwsh.bat")
		if fileExists(pwshBat) {
			return "cmd.exe", []string{"/C", pwshBat, "-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
		}
	}

	windir := os.Getenv("WINDIR")
	if windir == "" {
		windir = `C:\Windows`
	}
	return filepath.Join(windir, "System32", "WindowsPowerShell", "v1.0", "powershell.exe"),
		[]string{"-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func windowsFormScript(exePath string) string {
	return fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$exe = %s

$form = New-Object System.Windows.Forms.Form
$form.Text = 'JSON 转 CSV/Excel'
$form.StartPosition = 'CenterScreen'
$form.Width = 720
$form.Height = 330
$form.FormBorderStyle = [System.Windows.Forms.FormBorderStyle]::FixedDialog
$form.MaximizeBox = $false

$font = New-Object System.Drawing.Font('Microsoft YaHei UI', 9)
$form.Font = $font

$inputLabel = New-Object System.Windows.Forms.Label
$inputLabel.Text = 'JSON 文件'
$inputLabel.SetBounds(20, 22, 120, 24)
$form.Controls.Add($inputLabel)

$inputBox = New-Object System.Windows.Forms.TextBox
$inputBox.SetBounds(20, 48, 540, 28)
$form.Controls.Add($inputBox)

$inputButton = New-Object System.Windows.Forms.Button
$inputButton.Text = '浏览...'
$inputButton.SetBounds(575, 47, 100, 30)
$form.Controls.Add($inputButton)

$outputLabel = New-Object System.Windows.Forms.Label
$outputLabel.Text = '输出文件'
$outputLabel.SetBounds(20, 92, 120, 24)
$form.Controls.Add($outputLabel)

$outputBox = New-Object System.Windows.Forms.TextBox
$outputBox.SetBounds(20, 118, 540, 28)
$form.Controls.Add($outputBox)

$outputButton = New-Object System.Windows.Forms.Button
$outputButton.Text = '保存到...'
$outputButton.SetBounds(575, 117, 100, 30)
$form.Controls.Add($outputButton)

$formatLabel = New-Object System.Windows.Forms.Label
$formatLabel.Text = '输出格式'
$formatLabel.SetBounds(20, 162, 120, 24)
$form.Controls.Add($formatLabel)

$formatBox = New-Object System.Windows.Forms.ComboBox
$formatBox.DropDownStyle = [System.Windows.Forms.ComboBoxStyle]::DropDownList
[void]$formatBox.Items.Add('自动：默认 CSV')
[void]$formatBox.Items.Add('CSV')
[void]$formatBox.Items.Add('Excel XLSX')
$formatBox.SelectedIndex = 0
$formatBox.SetBounds(20, 188, 210, 28)
$form.Controls.Add($formatBox)

$convertButton = New-Object System.Windows.Forms.Button
$convertButton.Text = '转换'
$convertButton.SetBounds(20, 236, 120, 34)
$form.Controls.Add($convertButton)

$statusLabel = New-Object System.Windows.Forms.Label
$statusLabel.Text = ''
$statusLabel.SetBounds(160, 242, 500, 28)
$form.Controls.Add($statusLabel)

function Get-DefaultOutputPath {
    param([string]$inputPath, [string]$formatText)
    if ([string]::IsNullOrWhiteSpace($inputPath)) {
        return ''
    }
    $dir = [System.IO.Path]::GetDirectoryName($inputPath)
    $name = [System.IO.Path]::GetFileNameWithoutExtension($inputPath)
    if ([string]::IsNullOrWhiteSpace($name)) {
        $name = 'output'
    }
    $ext = '.csv'
    if ($formatText -eq 'Excel XLSX') {
        $ext = '.xlsx'
    }
    return [System.IO.Path]::Combine($dir, $name + $ext)
}

function Get-FormatArg {
    if ($formatBox.SelectedItem -eq 'CSV') {
        return 'csv'
    }
    if ($formatBox.SelectedItem -eq 'Excel XLSX') {
        return 'xlsx'
    }
    return ''
}

$inputButton.Add_Click({
    $dialog = New-Object System.Windows.Forms.OpenFileDialog
    $dialog.Title = '选择 JSON 文件'
    $dialog.Filter = 'JSON 文件 (*.json)|*.json|所有文件 (*.*)|*.*'
    if ($dialog.ShowDialog($form) -eq [System.Windows.Forms.DialogResult]::OK) {
        $inputBox.Text = $dialog.FileName
        if ([string]::IsNullOrWhiteSpace($outputBox.Text)) {
            $outputBox.Text = Get-DefaultOutputPath $dialog.FileName $formatBox.SelectedItem
        }
    }
})

$outputButton.Add_Click({
    $dialog = New-Object System.Windows.Forms.SaveFileDialog
    $dialog.Title = '保存转换结果'
    $dialog.Filter = 'CSV 文件 (*.csv)|*.csv|Excel 工作簿 (*.xlsx)|*.xlsx|所有文件 (*.*)|*.*'
    $dialog.AddExtension = $true
    $dialog.OverwritePrompt = $true
    $defaultPath = $outputBox.Text
    if ([string]::IsNullOrWhiteSpace($defaultPath)) {
        $defaultPath = Get-DefaultOutputPath $inputBox.Text $formatBox.SelectedItem
    }
    if (-not [string]::IsNullOrWhiteSpace($defaultPath)) {
        $dialog.FileName = $defaultPath
    }
    if ($dialog.ShowDialog($form) -eq [System.Windows.Forms.DialogResult]::OK) {
        $outputBox.Text = $dialog.FileName
    }
})

$formatBox.Add_SelectedIndexChanged({
    if (-not [string]::IsNullOrWhiteSpace($inputBox.Text)) {
        $outputBox.Text = Get-DefaultOutputPath $inputBox.Text $formatBox.SelectedItem
    }
})

$convertButton.Add_Click({
    $statusLabel.Text = ''
    if ([string]::IsNullOrWhiteSpace($inputBox.Text)) {
        [System.Windows.Forms.MessageBox]::Show($form, '请选择 JSON 文件。', '提示', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Warning) | Out-Null
        return
    }
    if ([string]::IsNullOrWhiteSpace($outputBox.Text)) {
        $outputBox.Text = Get-DefaultOutputPath $inputBox.Text $formatBox.SelectedItem
    }

    $cliArgs = @('-input', $inputBox.Text, '-output', $outputBox.Text)
    $formatArg = Get-FormatArg
    if (-not [string]::IsNullOrWhiteSpace($formatArg)) {
        $cliArgs += @('-format', $formatArg)
    }

    $result = & $exe @cliArgs 2>&1
    if ($LASTEXITCODE -ne 0) {
        [System.Windows.Forms.MessageBox]::Show($form, ($result | Out-String), '转换失败', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Error) | Out-Null
        return
    }
    $statusLabel.Text = '已保存到 ' + $outputBox.Text
    [System.Windows.Forms.MessageBox]::Show($form, '已保存到' + [Environment]::NewLine + $outputBox.Text, '转换完成', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information) | Out-Null
})

[void]$form.ShowDialog()
`, powerShellSingleQuoted(exePath))
}

func powerShellSingleQuoted(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
