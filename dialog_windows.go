package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Hilfsfunktion zum Abrufen des aktuellen Arbeitsverzeichnisses
func getCurrentDir() string {
	currentDir, err := os.Getwd()
	if err != nil || currentDir == "" {
		currentDir = "."
	}
	absDir, _ := filepath.Abs(currentDir)
	return absDir
}

// Nativer SaveFileDialog mit anpassbarem Filter und Dateinamen
func openNativeSaveDialog(defaultFilename, filter string) (string, bool) {
	absDir := getCurrentDir()

	if filter == "" {
		filter = "Alle Dateien (*.*)|*.*"
	}

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.SaveFileDialog
$dialog.InitialDirectory = '%s'
$dialog.FileName = '%s'
$dialog.Filter = '%s'
$dialog.Title = 'Datei speichern unter'
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    Write-Output $dialog.FileName
}
`, strings.ReplaceAll(absDir, "'", "''"), strings.ReplaceAll(defaultFilename, "'", "''"), strings.ReplaceAll(filter, "'", "''"))

	return executePowerShellDialog(psScript)
}

// Nativer OpenFileDialog (.car-Datei auswählen)
func openNativeOpenFileDialog() (string, bool) {
	absDir := getCurrentDir()

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.InitialDirectory = '%s'
$dialog.Filter = 'CAR-Dateien (*.car)|*.car|Alle Dateien (*.*)|*.*'
$dialog.Title = '.car-Datei auswählen'
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    Write-Output $dialog.FileName
}
`, strings.ReplaceAll(absDir, "'", "''"))

	return executePowerShellDialog(psScript)
}

// Nativer FolderBrowserDialog (Zielordner auswählen)
func openNativeFolderDialog() (string, bool) {
	absDir := getCurrentDir()

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.FolderBrowserDialog
$dialog.SelectedPath = '%s'
$dialog.Description = 'Zielordner für den Export auswählen'
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    Write-Output $dialog.SelectedPath
}
`, strings.ReplaceAll(absDir, "'", "''"))

	return executePowerShellDialog(psScript)
}

// Gemeinsamer, lautloser Ausführungs-Helper
func executePowerShellDialog(psScript string) (string, bool) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW

	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return "", false
	}

	return result, true
}