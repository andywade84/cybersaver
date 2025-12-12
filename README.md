# CyberSaver

CyberSaver is a Windows tray app that manages multiple Cyberpunk 2077 savegame profiles using junctions. It gives you a browser UI (served locally) to switch profiles, import/export saves, view quest info, and keep your saves safe.

## Features
- Profile management with junction switching (no game config needed).
- Auto-import current saves into a profile; per-profile notes.
- Save browser with screenshots, quest title/objective lookup, playtime, level, filters (auto/manual) and search.
- Copy saves between profiles; export profile as a ZIP.
- Safety prompts and automatic backup of the original save folder on first run; warnings if an existing junction points elsewhere.
- System tray presence with custom icon; reopen UI or exit from the tray.

## Requirements
- Windows with Cyberpunk 2077 saves in the default location (`%USERPROFILE%\Saved Games\CD Projekt Red\Cyberpunk 2077`) or manually selected.
- Go 1.22+ to build from source.

## Build
```powershell
go build
```
This produces `cybersaver.exe`. The web UI and quest data are embedded in the binary.

## Run
```powershell
.\cybersaver.exe
```
- First run shows safety prompts and creates a backup of the current save folder.
- The app opens `http://localhost:8787` in your browser. Close the browser without exiting the app; reopen from the system tray.
- Tray menu: “Open CyberSaver” (opens UI) and “Exit” (shuts down the server).

## Notes
- Profiles are stored under `profiles/` next to the executable. The game save folder is replaced with a junction to the selected profile when loaded.
- If you already have a junction at the game save path, CyberSaver warns before replacing it.
- The UI auto-refreshes saves every few seconds; use filters/search to narrow results.
