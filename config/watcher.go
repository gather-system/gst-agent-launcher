package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigPath returns the path of the config file currently in use, or empty if using embedded default.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err == nil {
		userPath := filepath.Join(home, ".config", "gst-launcher", "agents.json")
		if _, err := os.Stat(userPath); err == nil {
			return userPath
		}
	}

	exePath, err := os.Executable()
	if err == nil {
		localPath := filepath.Join(filepath.Dir(exePath), "agents.json")
		if _, err := os.Stat(localPath); err == nil {
			return localPath
		}
	}

	return ""
}

// WatchConfig watches the config file for changes and sends reloaded configs on the channel.
// It debounces events by 300ms to handle multiple rapid file system events on Windows.
func WatchConfig(path string) (<-chan *Config, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(filepath.Dir(path)); err != nil {
		watcher.Close()
		return nil, err
	}

	ch := make(chan *Config, 1)
	basename := filepath.Base(path)

	go func() {
		defer watcher.Close()
		var debounce *time.Timer

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(event.Name) != basename {
					continue
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(300*time.Millisecond, func() {
					cfg, err := loadFromFile(path)
					if err == nil {
						ch <- cfg
					}
				})
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return ch, nil
}
