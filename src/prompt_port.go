package main

import (
	"os/exec"
	"strconv"
	"strings"
)

// promptPort uses a Windows InputBox via PowerShell to ask the user for a port.
// Falls back to defaultPort if parsing fails.
func promptPort() int {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"[void][Reflection.Assembly]::LoadWithPartialName('Microsoft.VisualBasic');"+
			"$p=[Microsoft.VisualBasic.Interaction]::InputBox('Enter port for CyberSaver (1024-65535):','CyberSaver Port','"+strconv.Itoa(defaultPort)+"');"+
			"Write-Output $p")
	out, err := cmd.Output()
	if err != nil {
		return defaultPort
	}
	p := strings.TrimSpace(string(out))
	if p == "" {
		return defaultPort
	}
	val, err := strconv.Atoi(p)
	if err != nil || val < 1024 || val > 65535 {
		return defaultPort
	}
	return val
}
