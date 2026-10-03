//go:build windows

package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"
)

func writePNGFile(path string, img image.Image) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".lumix-export-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	// Never truncate the user's existing file before a complete replacement is ready.
	return os.Rename(name, path)
}

func preferredReportDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		if home, err := os.UserHomeDir(); err == nil {
			base = filepath.Join(home, "AppData", "Local")
		}
	}
	if base == "" {
		base = os.TempDir()
	}
	// Documents/Desktop/Pictures can be protected by Windows Controlled Folder Access.
	// Do not disable security or silently fall back to a temporary report location.
	return filepath.Join(base, "Lumerian", "LUMIX Usage Info", "Reports")
}
func writeUniqueReport(dir string, data []byte, now time.Time) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	stem := "LUMIX_Usage_Report_" + now.Format("20060102_150405")
	for i := 0; i < 10000; i++ {
		name := stem
		if i > 0 {
			name += fmt.Sprintf("_%d", i)
		}
		path := filepath.Join(dir, name+".txt")
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		n, writeErr := f.Write(data)
		closeErr := f.Close()
		if writeErr != nil || n != len(data) || closeErr != nil {
			os.Remove(path)
			if writeErr != nil {
				return "", writeErr
			}
			if closeErr != nil {
				return "", closeErr
			}
			return "", fmt.Errorf("incomplete report write")
		}
		return path, nil
	}
	return "", fmt.Errorf("too many report filename collisions")
}
