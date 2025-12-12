package main

import (
	"os"
	"path/filepath"
	"strings"
)

func defaultGameSavePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Saved Games", "CD Projekt Red", "Cyberpunk 2077")
}

func defaultProfilesDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "profiles"
	}
	return filepath.Join(filepath.Dir(exe), "profiles")
}

func detectGameSavePath() (string, bool) {
	p := defaultGameSavePath()
	return p, dirExists(p)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func samePath(a, b string) bool {
	aa, _ := filepath.Abs(a)
	bb, _ := filepath.Abs(b)
	return strings.EqualFold(aa, bb)
}
