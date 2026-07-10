package services

import (
	"context"
	"path/filepath"
	"time"

	"core/services/config"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// reloadDebounce is how long WatchDatabasesFile waits after the last filesystem
// event before reloading, so the burst of write/chmod/rename events a single
// save produces coalesces into one reload.
const reloadDebounce = 200 * time.Millisecond

// WatchDatabasesFile watches the databases config file and hot-reloads the engine
// whenever it changes, so edits to databases.yaml take effect without a restart.
//
// It watches the file's *directory* rather than the file itself, so it survives
// atomic saves (write-temp-then-rename, what editors and bind-mount updates do)
// and picks up a file that is created after startup. Events are debounced, then
// the path is re-resolved (handling a .yaml/.yml switch) and the engine reloaded;
// a parse/validation error is logged and the previous set is kept.
//
// It blocks until ctx is cancelled, returning nil on graceful shutdown, so it
// slots into the serve errgroup alongside the HTTP servers. If a watcher can't be
// established it logs and returns nil — hot-reload degrades off, the server runs.
func (s *Services) WatchDatabasesFile(ctx context.Context) error {
	configured := s.Config.DatabasesFile
	directory := filepath.Dir(configured)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Warn().Err(err).Msg("hot-reload disabled: cannot create fsnotify watcher")
		return nil
	}
	defer watcher.Close()

	if err := watcher.Add(directory); err != nil {
		log.Warn().Err(err).Str("directory", directory).Msg("hot-reload disabled: cannot watch directory")
		return nil
	}
	log.Info().Str("file", configured).Msg("watching databases file for changes")

	// A single shared timer debounces the event burst. It starts stopped; the
	// first relevant event arms it, and its firing triggers the reload.
	debounce := time.NewTimer(reloadDebounce)
	debounce.Stop()
	defer debounce.Stop()

	target := filepath.Base(configured)
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// Only react to events on our file (any extension we might resolve to),
			// not to unrelated files sharing the directory.
			if relevantEvent(event.Name, target) {
				debounce.Reset(reloadDebounce)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Error().Err(err).Msg("databases file watch error")
		case <-debounce.C:
			s.reloadDatabases(configured)
		}
	}
}

// relevantEvent reports whether an event on eventPath concerns the databases
// config. It matches the configured base name and its .yaml/.yml sibling, so a
// switch between the two extensions still triggers a reload.
func relevantEvent(eventPath, target string) bool {
	name := filepath.Base(eventPath)
	if name == target {
		return true
	}
	base := target[:len(target)-len(filepath.Ext(target))]
	return name == base+".yaml" || name == base+".yml"
}

// reloadDatabases re-resolves the configured path and reloads the engine, logging
// the outcome. A resolve or parse/validation failure keeps the previous set.
func (s *Services) reloadDatabases(configured string) {
	path, err := config.ResolveDatabasesFile(configured)
	if err != nil {
		log.Warn().Err(err).Msg("databases reload failed (resolve), keeping previous set")
		return
	}
	if err := s.Engine.Reload(path); err != nil {
		log.Warn().Err(err).Msg("databases reload failed (parse/validate), keeping previous set")
		return
	}
	// A reload is the one moment a name's connection details can change, so drop
	// the cached name->database_id mappings and let the next query re-derive.
	s.Databases.InvalidateCache()
	log.Info().Int("count", len(s.Engine.List())).Str("path", path).Msg("databases reloaded")
}
