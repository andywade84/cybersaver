package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func classifySave(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "auto"):
		return "Auto"
	case strings.Contains(n, "manual"):
		return "Manual"
	default:
		return "Other"
	}
}

func findScreenshot(path string) string {
	candidates := []string{"screenshot.png", "screenshot.jpg", "screenshot.jpeg", "screenshot.bmp"}
	for _, c := range candidates {
		fp := filepath.Join(path, c)
		if _, err := os.Stat(fp); err == nil {
			return filepath.ToSlash(filepath.Join(filepath.Base(path), c))
		}
	}
	return ""
}

func readMetadata(saveDir string) metaSummary {
	files, _ := filepath.Glob(filepath.Join(saveDir, "metadata*.json"))
	if len(files) == 0 {
		return metaSummary{}
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		return metaSummary{}
	}
	var meta struct {
		Data struct {
			Metadata struct {
				TrackedQuestEntry string  `json:"trackedQuestEntry"`
				PlayTime          float64 `json:"playTime"`
				Level             float64 `json:"level"`
			} `json:"metadata"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return metaSummary{}
	}
	qTitle, obj := quests.lookup(meta.Data.Metadata.TrackedQuestEntry)
	return metaSummary{
		Playtime:   formatPlaytime(meta.Data.Metadata.PlayTime),
		Level:      formatLevel(meta.Data.Metadata.Level),
		Quest:      trimQuest(meta.Data.Metadata.TrackedQuestEntry),
		QuestTitle: qTitle,
		Objective:  obj,
	}
}

func formatPlaytime(seconds float64) string {
	if seconds <= 0 {
		return ""
	}
	d := time.Duration(seconds * float64(time.Second))
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

func formatLevel(level float64) string {
	if level <= 0 {
		return ""
	}
	return fmt.Sprintf("Lvl %d", int(level))
}

func trimQuest(q string) string {
	if q == "" {
		return ""
	}
	if idx := strings.LastIndex(q, "/"); idx >= 0 && idx < len(q)-1 {
		return q[idx+1:]
	}
	return q
}
