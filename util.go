package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func sanitizeName(in string) string {
	s := strings.TrimSpace(in)
	s = strings.ReplaceAll(s, string(filepath.Separator), "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func openBrowser(url string) {
	// best-effort for Windows; ignore errors
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func readNote(profilesDir, profile string) string {
	path := filepath.Join(profilesDir, profile, ".note.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeNote(profilesDir, profile, note string) error {
	path := filepath.Join(profilesDir, profile, ".note.txt")
	return os.WriteFile(path, []byte(note), 0o644)
}
