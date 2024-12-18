package collector

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nilszeilon/devstats/internal/domain"
	"github.com/nilszeilon/devstats/internal/storage"
)

const maxWatchedDirs = 1000 // Adjust this number based on your needs

type FileChangeCollector struct {
	store    storage.Store[domain.FileChangeData]
	watcher  *fsnotify.Watcher
	stopChan chan struct{}
	paths    []string
}

func NewFileChangeCollector(store storage.Store[domain.FileChangeData], paths []string) (*FileChangeCollector, error) {
	// Increase system file descriptor limit
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return nil, fmt.Errorf("error getting rlimit: %v", err)
	}

	// Set to a higher value, but keep it under the system maximum
	newLimit := syscall.Rlimit{
		Cur: 10240, // Soft limit
		Max: rLimit.Max,
	}
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &newLimit)
	if err != nil {
		log.Printf("Warning: Could not increase file descriptor limit: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileChangeCollector{
		store:    store,
		watcher:  watcher,
		stopChan: make(chan struct{}),
		paths:    paths,
	}, nil
}

func (fc *FileChangeCollector) Start() error {
	watchedDirs := 0
	// Add paths to watch
	for _, path := range fc.paths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			// Handle permission errors and other access issues
			if err != nil {
				// log.Printf("Error accessing path %s: %v", path, err)
				return filepath.SkipDir
			}

			if info.IsDir() {
				base := filepath.Base(path)
				// Skip hidden directories (starting with a dot)
				if len(base) > 0 && base[0] == '.' {
					// log.Printf("Skipping hidden directory: %s", path)
					return filepath.SkipDir
				}

				// Skip blacklisted directories
				if isBlacklistedDir(path) {
					// log.Printf("Skipping blacklisted directory: %s", path)
					return filepath.SkipDir
				}

				// Check if we've hit the watch limit
				if watchedDirs >= maxWatchedDirs {
					log.Printf("Reached maximum number of watched directories (%d), skipping: %s", maxWatchedDirs, path)
					return filepath.SkipDir
				}

				// Try to add the directory to the watcher
				if err := fc.watcher.Add(path); err != nil {
					log.Printf("Error watching directory %s: %v", path, err)
					return filepath.SkipDir
				}
				watchedDirs++
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error walking path %s: %v", path, err)
		}
	}

	go fc.watch()
	return nil
}

func (fc *FileChangeCollector) watch() {
	for {
		select {
		case <-fc.stopChan:
			return
		case event, ok := <-fc.watcher.Events:
			if !ok {
				return
			}

			// Skip non-code files (you might want to customize this)
			if !isCodeFile(event.Name) {
				continue
			}

			switch {
			case event.Op&fsnotify.Write == fsnotify.Write:
			case event.Op&fsnotify.Create == fsnotify.Create:
			case event.Op&fsnotify.Remove == fsnotify.Remove:
			default:
				// we don't want chmod changes
				continue
			}

			language := getLanguage(event.Name)
			if language == "" {
				continue
			}

			data := domain.FileChangeData{
				Language:  language,
				Timestamp: time.Now(),
			}

			if err := fc.store.Save(data); err != nil {
				log.Printf("Error saving file change: %v", err)
			}

		case err, ok := <-fc.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (fc *FileChangeCollector) Stop() {
	close(fc.stopChan)
	fc.watcher.Close()
}

// isBlacklistedDir returns true if the directory should be skipped
func isBlacklistedDir(path string) bool {
	base := filepath.Base(path)
	blacklist := map[string]bool{
		// macOS system directories
		"Library":      true,
		"Applications": true,
		"System":       true,
		"Volumes":      true,
		"cores":        true,
		"private":      true,

		// Development related directories to skip
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		"target":       true,
		"coverage":     true,
		"tmp":          true,
		"temp":         true,
		"go":           true,
		"rails":        true,

		// Package manager directories
		"bower_components": true,
		"jspm_packages":    true,
		"packages":         true,

		// IDE and editor directories
		".idea":     true,
		".vscode":   true,
		".eclipse":  true,
		".settings": true,

		// Version control
		".git": true,
		".svn": true,
		".hg":  true,

		// macOS specific
		".Trash": true,
		".cache": true,
		".npm":   true,
		".yarn":  true,
	}
	return blacklist[base]
}

func getLanguage(path string) string {
	ext := filepath.Ext(path)
	languageMap := map[string]string{
		".go":     "go",
		".js":     "javascript",
		".ts":     "typescript",
		".svelte": "svelte",
		".py":     "python",
		".rb":     "ruby",
		".md":     "markdown",
		".java":   "java",
		".c":      "c",
		".rs":     "rust",
		".css":    "css",
		".html":   "html",
		".sql":    "sql",
		".sh":     "shell",
		".yaml":   "yaml",
		".yml":    "yaml",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	return ""
}

func isCodeFile(path string) bool {
	return getLanguage(path) != ""
}
