// Command desktop is the Wails v2 entry point that packages jobtracker as a
// native desktop app: the frontend is embedded, the store lives in the OS
// user data dir, and there is no basic auth or network listener (FR-02,
// FR-03). It shares handler.Routes and store.Store with the web target
// unchanged (main.go at the repo root).
package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tecnologer/jobtracker/handler"
	"github.com/tecnologer/jobtracker/store"
	"github.com/tecnologer/jobtracker/web"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		var err error
		dbPath, err = defaultDBPath()
		if err != nil {
			log.Fatal(err)
		}
	}

	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	dist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		log.Fatalf("preparing embedded frontend: %v", err)
	}
	mux := handler.Routes(handler.New(s), http.FileServer(http.FS(dist)))

	err = wails.Run(&options.App{
		Title:  "JobTracker",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Handler: mux,
		},
		OnShutdown: func(_ context.Context) {
			if err := s.Close(); err != nil {
				log.Printf("closing store: %v", err)
			}
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// defaultDBPath returns the per-user data directory for jobtracker's SQLite
// file, creating the directory if it doesn't exist yet. On Linux it honors
// XDG_DATA_HOME (falling back to ~/.local/share); on other OSes it uses
// os.UserConfigDir, which resolves to %AppData% on Windows and
// ~/Library/Application Support on macOS.
func defaultDBPath() (string, error) {
	base, err := userDataDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "jobtracker")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating data directory: %w", err)
	}
	return filepath.Join(dir, "jobs.db"), nil
}

// userDataDir resolves the OS-appropriate per-user data directory: XDG_DATA_HOME
// (falling back to ~/.local/share) on Linux, os.UserConfigDir elsewhere (which
// resolves to %AppData% on Windows and ~/Library/Application Support on macOS).
func userDataDir() (string, error) {
	if runtime.GOOS != "linux" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolving user config directory: %w", err)
		}
		return dir, nil
	}

	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share"), nil
}
