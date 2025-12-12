package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sqweek/dialog"
)

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := webFS.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "UI missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	profiles := s.listProfiles()
	active := s.detectActiveProfile(profiles)
	path := ""
	if s.gamePathExists && dirExists(s.gameSavePath) {
		path = s.gameSavePath
	}
	writeJSON(w, map[string]any{
		"profiles":    profiles,
		"active":      active,
		"gamePath":    path,
		"pathMissing": path == "",
		"profilesDir": s.profilesDir,
	})
}

func (s *server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.listProfiles())
	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		name := sanitizeName(body.Name)
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		path := filepath.Join(s.profilesDir, name)
		if _, err := os.Stat(path); err == nil {
			http.Error(w, "profile exists", http.StatusConflict)
			return
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"status": "created"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) handleProfileNote(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		profile := sanitizeName(r.URL.Query().Get("profile"))
		if profile == "" {
			http.Error(w, "profile required", http.StatusBadRequest)
			return
		}
		writeJSON(w, profileNote{Profile: profile, Note: readNote(s.profilesDir, profile)})
	case http.MethodPost:
		var body profileNote
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if body.Profile == "" {
			http.Error(w, "profile required", http.StatusBadRequest)
			return
		}
		if err := writeNote(s.profilesDir, body.Profile, body.Note); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"status": "saved"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) handleProfileDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	name = sanitizeName(name)
	if name == "" {
		http.Error(w, "invalid profile", http.StatusBadRequest)
		return
	}
	target := filepath.Join(s.profilesDir, name)
	if target == "" {
		http.NotFound(w, r)
		return
	}
	if link, _ := os.Readlink(s.gameSavePath); samePath(link, target) {
		http.Error(w, "profile is active, unload first", http.StatusConflict)
		return
	}
	if err := os.RemoveAll(target); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "deleted"})
}

func (s *server) handleLoadProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.gameSavePath == "" {
		http.Error(w, "game save path not set", http.StatusBadRequest)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	name := sanitizeName(body.Name)
	if name == "" {
		http.Error(w, "invalid profile", http.StatusBadRequest)
		return
	}
	target := filepath.Join(s.profilesDir, name)
	if err := os.MkdirAll(target, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := switchJunction(s.gameSavePath, target); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "loaded", "profile": name})
}

func (s *server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.gameSavePath == "" {
		http.Error(w, "game save path not set", http.StatusBadRequest)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	name := sanitizeName(body.Name)
	if name == "" {
		http.Error(w, "invalid profile", http.StatusBadRequest)
		return
	}
	if err := s.importFromGamePath(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "imported"})
}

func (s *server) handleSaves(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	profile := sanitizeName(r.URL.Query().Get("profile"))
	if profile == "" {
		http.Error(w, "profile required", http.StatusBadRequest)
		return
	}
	base := filepath.Join(s.profilesDir, profile)
	entries, err := os.ReadDir(base)
	if err != nil {
		writeJSON(w, []saveInfo{})
		return
	}
	type saveWithTime struct {
		saveInfo
		mod time.Time
	}
	var saves []saveWithTime
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		savePath := filepath.Join(base, e.Name())
		ss := findScreenshot(savePath)
		meta := readMetadata(savePath)
		saves = append(saves, saveWithTime{
			saveInfo: saveInfo{
				Name:       e.Name(),
				Modified:   info.ModTime().Format("2006-01-02 15:04:05"),
				Type:       classifySave(e.Name()),
				Screenshot: ss,
				Playtime:   meta.Playtime,
				Level:      meta.Level,
				Quest:      meta.Quest,
				QuestTitle: meta.QuestTitle,
				Objective:  meta.Objective,
			},
			mod: info.ModTime(),
		})
	}
	sort.Slice(saves, func(i, j int) bool { return saves[i].mod.After(saves[j].mod) })
	resp := make([]saveInfo, 0, len(saves))
	for _, s := range saves {
		resp = append(resp, s.saveInfo)
	}
	writeJSON(w, resp)
}

func (s *server) handleDeleteSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Profile string `json:"profile"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	profile := sanitizeName(body.Profile)
	name := sanitizeName(body.Name)
	if profile == "" || name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	target := filepath.Join(s.profilesDir, profile, name)
	if err := os.RemoveAll(target); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "deleted"})
}

func (s *server) handleCopySave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Profile string `json:"profile"`
		Name    string `json:"name"`
		Target  string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	srcProfile := sanitizeName(body.Profile)
	name := sanitizeName(body.Name)
	targetProfile := sanitizeName(body.Target)
	if srcProfile == "" || name == "" || targetProfile == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	srcPath := filepath.Join(s.profilesDir, srcProfile, name)
	if _, err := os.Stat(srcPath); err != nil {
		http.Error(w, "source save not found", http.StatusNotFound)
		return
	}
	destDir := filepath.Join(s.profilesDir, targetProfile)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	destPath := filepath.Join(destDir, name)
	if _, err := os.Stat(destPath); err == nil {
		destPath = filepath.Join(destDir, name+"_copy_"+time.Now().Format("20060102_150405"))
	}
	if err := copyDir(srcPath, destPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "copied", "dest": destPath})
}

func (s *server) handleExportProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	profile := sanitizeName(r.URL.Query().Get("profile"))
	if profile == "" {
		http.Error(w, "profile required", http.StatusBadRequest)
		return
	}
	zipPath, err := createProfileZip(s.profilesDir, profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(zipPath)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+profile+".zip\"")
	http.ServeFile(w, r, zipPath)
}

func (s *server) handleSelectPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path, err := dialog.Directory().Title("Select Cyberpunk 2077 Save Folder").Browse()
	if err != nil {
		http.Error(w, "selection cancelled", http.StatusBadRequest)
		return
	}
	s.gameSavePath = path
	s.gamePathExists = dirExists(path)
	writeJSON(w, map[string]string{"path": path})
}

func (s *server) listProfiles() []string {
	entries, err := os.ReadDir(s.profilesDir)
	if err != nil {
		return []string{}
	}
	var res []string
	for _, e := range entries {
		if e.IsDir() {
			res = append(res, e.Name())
		}
	}
	return res
}

func (s *server) detectActiveProfile(profiles []string) string {
	target, err := os.Readlink(s.gameSavePath)
	if err != nil {
		return ""
	}
	for _, p := range profiles {
		if samePath(target, filepath.Join(s.profilesDir, p)) {
			return p
		}
	}
	return ""
}
