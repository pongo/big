//go:build windows

package fs

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func OpenPath(path string) error {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", path).Start()
}

func RevealPath(path string) error {
	return startExplorer(explorerRevealCommandLine(path))
}

func explorerSelectArgument(path string) string {
	return `/select,"` + filepath.Clean(path) + `"`
}

func explorerRevealCommandLine(path string) string {
	return explorerExecutable() + " " + explorerSelectArgument(path)
}

func explorerExecutable() string {
	windowsDir := os.Getenv("WINDIR")
	if windowsDir == "" {
		return "explorer.exe"
	}
	return `"` + filepath.Join(windowsDir, "explorer.exe") + `"`
}

func startExplorer(commandLine string) error {
	var startupInfo windows.StartupInfo
	var processInfo windows.ProcessInformation
	if err := windows.CreateProcess(
		nil,
		windows.StringToUTF16Ptr(commandLine),
		nil,
		nil,
		false,
		0,
		nil,
		nil,
		&startupInfo,
		&processInfo,
	); err != nil {
		return err
	}
	windows.CloseHandle(processInfo.Thread)
	windows.CloseHandle(processInfo.Process)
	return nil
}
