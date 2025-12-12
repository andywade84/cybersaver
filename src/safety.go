package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sqweek/dialog"
)

// ensureProtection prompts the user about junction usage and optionally backs up saves.
// It writes a marker file in the profiles directory so we only prompt once.
func ensureProtection(s *server) {
	marker := filepath.Join(s.profilesDir, ".warning_ack")
	if _, err := os.Stat(marker); err == nil {
		return
	}

	message := "CyberSaver swaps your Cyberpunk 2077 save folder with a junction. If you delete this CyberSaver folder, you could lose access to saves stored here.\n\nContinue and create a backup of your current saves?"
	confirmed := dialog.Message(message).Title("CyberSaver Warning").YesNo()
	if !confirmed {
		log.Printf("User cancelled at warning prompt; exiting.")
		os.Exit(0)
	}

	// Warn if an existing junction points somewhere else.
	if info, err := os.Lstat(s.gameSavePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		if target, err := os.Readlink(s.gameSavePath); err == nil {
			if !pointsIntoProfiles(target, s.profilesDir) {
				msg := "An existing junction/symlink is already at the game save location and points to:\n" + target + "\n\nCyberSaver will replace it and point to its profiles directory. Continue?"
				replace := dialog.Message(msg).Title("Replace Existing Junction").YesNo()
				if !replace {
					log.Printf("User cancelled at replace-junction prompt; exiting.")
					os.Exit(0)
				}
			}
		}
	}

	if s.gamePathExists && dirExists(s.gameSavePath) {
		backupDir := filepath.Join(filepath.Dir(s.gameSavePath), "Cyberpunk 2077_backup_"+time.Now().Format("20060102_150405"))
		if err := copyDir(s.gameSavePath, backupDir); err != nil {
			log.Printf("Backup failed: %v", err)
		} else {
			log.Printf("Backup created at %s", backupDir)
		}
	}

	_ = os.WriteFile(marker, []byte("ack"), 0o644)
}

func pointsIntoProfiles(target, profilesDir string) bool {
	t, err1 := filepath.Abs(target)
	p, err2 := filepath.Abs(profilesDir)
	if err1 != nil || err2 != nil {
		return false
	}
	t = strings.ToLower(filepath.Clean(t))
	p = strings.ToLower(filepath.Clean(p))
	return strings.HasPrefix(t, p)
}
