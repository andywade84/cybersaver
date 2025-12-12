package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func switchJunction(linkPath, target string) error {
	info, err := os.Lstat(linkPath)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(linkPath); err != nil {
				return err
			}
		} else if info.IsDir() {
			backup := linkPath + "_backup_" + time.Now().Format("20060102_150405")
			if err := os.Rename(linkPath, backup); err != nil {
				return fmt.Errorf("existing folder could not be moved: %w", err)
			}
		} else {
			return fmt.Errorf("existing path is not a folder")
		}
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("cmd", "/C", "mklink", "/J", linkPath, target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mklink failed: %v (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *server) importFromGamePath(profile string) error {
	if s.gameSavePath == "" {
		return fmt.Errorf("game save path not set")
	}
	src := s.gameSavePath
	dest := filepath.Join(s.profilesDir, profile)
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		destPath := filepath.Join(dest, e.Name())
		if err := copyDir(srcPath, destPath); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return copyFile(src, dest)
	}
	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyDir(filepath.Join(src, e.Name()), filepath.Join(dest, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dest string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}

func createProfileZip(profilesDir, profile string) (string, error) {
	base := filepath.Join(profilesDir, profile)
	if _, err := os.Stat(base); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp("", "profile_*.zip")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	zw := zip.NewWriter(tmp)
	err = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(filepath.Dir(base), path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			_, err = zw.Create(rel + "/")
			return err
		}
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
	if err != nil {
		zw.Close()
		return "", err
	}
	if err := zw.Close(); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}
