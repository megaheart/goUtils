package main

import (
	"errors"
	"path/filepath"

	"github.com/megaheart/goUtils/fs"
	"github.com/megaheart/goUtils/log"
)

func Test_FileWatcher() {
	logger := log.NewZapLogger(
		log.ZapLogger_Format_ReadableText,
		"2026-01-02 15:04:05.000",
		log.ZapLogger_Output_Console,
		"",
		log.ZapLogger_Level_Debug,
		nil,
	)
	watcher, err := fs.NewFileSystem().NewFileWatcher()
	if err != nil {
		panic(err)
	}

	absolutePath, err := filepath.Abs("dist/hi")
	if err != nil {
		panic(err)
	}
	parentDir := filepath.Dir(absolutePath)
	if parentDir == "" {
		panic(errors.New("Failed to get parent directory of " + absolutePath))
	}

	watcher.Add(parentDir)
	// watcher.Add(absolutePath)

	// logger.Info("Start watching file: " + absolutePath)
	logger.Info("Start watching parent directory: " + parentDir)

	for {
		select {
		case event, ok := <-watcher.Events():
			if !ok {
				logger.Error("Watcher stop watching file because fileWatcher.Events channel is closed (ok = false, panic call)")
				return
			}
			logger.Info("File modified: "+event.Path, log.LogString("event.Op", event.Op.String()))

		case err, ok := <-watcher.Errors():
			if !ok {
				if err == nil {
					err = errors.New("error=nil, watcher stop watching file because fileWatcher.Errors channel is closed")
				}
				logger.Error("Watcher stop watching file because fileWatcher.Errors channel is closed", log.LogError(err))
				// system.Fatal("Error watching file (ok = false, panic call)", 1)
				return
			}
			if err == nil {
				err = errors.New("error=nil, ok=true, the behavior of the watcher.Errors function is unknown")
			} else {
				logger.Error("Error watching file", log.LogError(err))
				// return
			}
		}
	}
}
