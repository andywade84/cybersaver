package main

import (
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var gameRunning atomic.Bool

func isGameRunning() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq Cyberpunk2077.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "cyberpunk2077.exe")
}

func monitorGameState(updateIcon func(running bool)) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	set := func(v bool) {
		if gameRunning.Load() != v {
			gameRunning.Store(v)
			if updateIcon != nil {
				updateIcon(v)
			}
		}
	}
	set(isGameRunning())
	for range ticker.C {
		set(isGameRunning())
	}
}
