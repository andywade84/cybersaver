package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"strings"
)

//go:embed journal-quest-data.json
var questData []byte

type questIndex struct {
	pathToTitle    map[string]string
	objectiveToDoc map[string]string
}

type questEntry struct {
	Path        string       `json:"path"`
	Title       string       `json:"title"`
	Phases      []questPhase `json:"phases"`
	Description string       `json:"description"`
}

type questPhase struct {
	Path       string           `json:"path"`
	Objectives []questObjective `json:"objectives"`
}

type questObjective struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

var quests = loadQuestIndex()

func loadQuestIndex() *questIndex {
	idx := &questIndex{
		pathToTitle:    map[string]string{},
		objectiveToDoc: map[string]string{},
	}
	if len(questData) == 0 {
		log.Printf("quest data not embedded; quest titles will be unavailable")
		return idx
	}
	var entries []questEntry
	if err := json.Unmarshal(questData, &entries); err != nil {
		log.Printf("failed to parse quest data: %v", err)
		return idx
	}
	for _, q := range entries {
		addPath(idx.pathToTitle, q.Path, q.Title)
		for _, ph := range q.Phases {
			addPath(idx.pathToTitle, ph.Path, q.Title)
			for _, obj := range ph.Objectives {
				addPath(idx.pathToTitle, obj.Path, q.Title)
				addPath(idx.objectiveToDoc, obj.Path, obj.Description)
			}
		}
	}
	return idx
}

func addPath(m map[string]string, path, value string) {
	if path == "" {
		return
	}
	n := normalizePath(path)
	if _, ok := m[n]; !ok {
		m[n] = value
	}
}

func normalizePath(p string) string {
	return strings.Trim(strings.ToLower(p), "/")
}

func parentPath(p string) string {
	if p == "" {
		return ""
	}
	if idx := strings.LastIndex(p, "/"); idx > 0 {
		return p[:idx]
	}
	return ""
}

func (q *questIndex) lookup(path string) (title string, objective string) {
	if q == nil {
		return "", ""
	}
	np := normalizePath(path)
	if np == "" {
		return "", ""
	}
	if doc, ok := q.objectiveToDoc[np]; ok {
		objective = doc
	}
	for cursor := np; cursor != ""; cursor = parentPath(cursor) {
		if t, ok := q.pathToTitle[cursor]; ok {
			title = t
			break
		}
	}
	return title, objective
}
