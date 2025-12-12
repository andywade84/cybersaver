package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getlantern/systray"
)

//go:generate rsrc -ico CyberSaver.ico -o CyberSaver.syso

const shutdownTimeout = 3 * time.Second
const defaultPort = 8787

func main() {
	autoPath, ok := detectGameSavePath()
	s := &server{
		gameSavePath:   autoPath,
		gamePathExists: ok,
		profilesDir:    defaultProfilesDir(),
	}

	cfg := requirePort(loadConfig())
	if cfg.GameSavePath != "" {
		s.gameSavePath = cfg.GameSavePath
		s.gamePathExists = dirExists(cfg.GameSavePath)
	}
	if cfg.ProfilesDir != "" {
		s.profilesDir = cfg.ProfilesDir
	}
	if err := os.MkdirAll(s.profilesDir, 0o755); err != nil {
		log.Fatalf("failed to create profiles dir: %v", err)
	}

	cfg = runSetupWizard(cfg, s)
	ensureProtection(s)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(s.profilesDir))))
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/profiles", s.handleProfiles)
	mux.HandleFunc("/api/profiles/", s.handleProfileDelete)
	mux.HandleFunc("/api/profile_note", s.handleProfileNote)
	mux.HandleFunc("/api/copy_save", s.handleCopySave)
	mux.HandleFunc("/api/export_profile", s.handleExportProfile)
	mux.HandleFunc("/api/load", s.handleLoadProfile)
	mux.HandleFunc("/api/import", s.handleImport)
	mux.HandleFunc("/api/saves", s.handleSaves)
	mux.HandleFunc("/api/delete_save", s.handleDeleteSave)
	mux.HandleFunc("/api/select_path", s.handleSelectPath)

	addr := "localhost:" + strconv.Itoa(configPort(cfg))
	url := "http://" + addr
	httpServer := &http.Server{Addr: addr, Handler: mux}

	go func() {
		log.Printf("CyberSaver running at %s", url)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	systray.Run(func() {
		systray.SetTitle("CyberSaver")
		systray.SetTooltip("Cyberpunk 2077 Save Profiles")
		if data := iconBytes(); len(data) > 0 {
			systray.SetIcon(data)
		}
		openItem := systray.AddMenuItem("Open CyberSaver", "Open UI in browser")
		exitItem := systray.AddMenuItem("Exit", "Quit CyberSaver")
		go openBrowser(url)
		go func() {
			for {
				select {
				case <-openItem.ClickedCh:
					openBrowser(url)
				case <-exitItem.ClickedCh:
					shutdownServer(httpServer)
					systray.Quit()
					return
				}
			}
		}()
		go monitorGameState(func(running bool) {
			if running {
				systray.SetIcon(iconBytesDanger())
				systray.SetTooltip("Cyberpunk running - switches blocked")
			} else {
				systray.SetIcon(iconBytes())
				systray.SetTooltip("Cyberpunk 2077 Save Profiles")
			}
		})
	}, func() {
		shutdownServer(httpServer)
	})
}

func shutdownServer(s *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	_ = s.Shutdown(ctx)
}
