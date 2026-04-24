package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/megaheart/goUtils/fs"
	fsUtils "github.com/megaheart/goUtils/fs/utils"
	"github.com/megaheart/goUtils/log"
)

func Test_OsFileWatcher() {
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
				return
			}
			if err == nil {
				err = errors.New("error=nil, ok=true, the behavior of the watcher.Errors function is unknown")
			} else {
				logger.Error("Error watching file", log.LogError(err))
			}
		}
	}
}

func Test_FileWatcher_LogRotate() {
	path, err := filepath.Abs("dist/hi")
	if err != nil {
		panic(err)
	}

	go func() {
		// Simulate log rotation by renaming the file every 5 seconds
		// Print 1 line log per second to the file
		fs := fs.NewFileSystem()
		distDir := filepath.Dir(path)
		if err := fs.MakeDirAll(distDir, 0755); err != nil {
			panic(err)
		}

		lineTicker := time.NewTicker(1 * time.Second)
		rotateTicker := time.NewTicker(5 * time.Second)
		defer lineTicker.Stop()
		defer rotateTicker.Stop()

		var lineNo int64 = 0
		for {
			select {
			case <-lineTicker.C:
				lineNo++
				line := fmt.Sprintf("[%s] line=%d simulated log\n", time.Now().Format("15:04:05.000"), lineNo)
				if _, err := fs.AppendFile(path, []byte(line)); err != nil {
					fmt.Printf("[SIM][%s] append log error: %v\n", time.Now().Format("15:04:05.000"), err)
				}
			case <-rotateTicker.C:
				rotatedPath := fmt.Sprintf("%s.%d", path, time.Now().UnixNano())
				if err := fs.MoveFile(path, rotatedPath); err != nil {
					if !fs.IsNotExistError(err) {
						fmt.Printf("[SIM][%s] rotate error: %v\n", time.Now().Format("15:04:05.000"), err)
					}
					continue
				}
				fmt.Printf("[SIM][%s] rotated => %s\n", time.Now().Format("15:04:05.000"), rotatedPath)
			}
		}
	}()

	go func() {
		logger := log.NewZapLogger(
			log.ZapLogger_Format_ReadableText,
			"2026-01-02 15:04:05.000",
			log.ZapLogger_Output_Console,
			"",
			log.ZapLogger_Level_Debug,
			nil,
		)
		// Read log by github.com/nxadm/tail + fs_utils.FileWatcher, and print log content to console
		fs := fs.NewFileSystem()
		watcher := fsUtils.NewFileWatcher(logger, fs, 200*time.Millisecond)
		defer watcher.Close()

		// doNewest := concurrency.NewDoNewest()
		skipOld := true
		watcher.WatchFile(path, fsUtils.FileWatchMode_Replace, func() error {
			lines, err := fsUtils.TailFilePolling_DependOnINode(path, skipOld, 200*time.Millisecond, 10*time.Second)
			if err != nil {
				fmt.Printf("[TAIL][%s] error initializing tail: %v\n", time.Now().Format("15:04:05.000"), err)
				return err
			}
			skipOld = false // only skip old logs for the first time, after that always read all logs in the file, even after log rotation, to avoid missing any log line
			go func() {
				for line := range lines {
					if line.Err != nil {
						fmt.Printf("[TAIL][%s] error: %v\n", time.Now().Format("15:04:05.000"), line.Err)
						continue
					}
					fmt.Printf("[TAIL][%s] %s", time.Now().Format("15:04:05.000"), line.Line)
				}
				fmt.Printf("[TAIL][%s] stop tailing file (channel closed)\n", time.Now().Format("15:04:05.000"))
			}()
			return nil
		})
	}()

	// Press Ctrl+C to stop the test after observing the log rotation behavior for a while
	select {}
}

func Test_FileWatcher_FileModify() {
	logger := log.NewZapLogger(
		log.ZapLogger_Format_ReadableText,
		"2026-01-02 15:04:05.000",
		log.ZapLogger_Output_Console,
		"",
		log.ZapLogger_Level_Debug,
		nil,
	)

	osFs := fs.NewFileSystem()
	if err := osFs.MakeDirAll("dist", 0755); err != nil {
		panic(err)
	}
	targetPath, err := filepath.Abs("dist/hi")
	if err != nil {
		panic(err)
	}
	tempRenamePath := targetPath + ".tmp"
	_ = os.Remove(tempRenamePath)

	if err := os.WriteFile(targetPath, []byte("v1: initial content\n"), 0644); err != nil {
		panic(err)
	}

	fw := fsUtils.NewFileWatcher(logger, osFs, 200*time.Millisecond)
	defer fw.Close()

	var mu sync.Mutex
	onChangeTimes := make([]time.Time, 0)
	onChange := func() error {
		mu.Lock()
		now := time.Now()
		onChangeTimes = append(onChangeTimes, now)
		idx := len(onChangeTimes)
		mu.Unlock()

		data, readErr := os.ReadFile(targetPath)
		if readErr != nil {
			fmt.Printf("[onChange #%d][%s] read dist/hi error: %v\n", idx, now.Format("15:04:05.000"), readErr)
			return nil
		}
		fmt.Printf("[onChange #%d][%s] dist/hi content:\n%s", idx, now.Format("15:04:05.000"), string(data))
		return nil
	}

	if err := fw.WatchFile(targetPath, fsUtils.FileWatchMode_Modify, onChange); err != nil {
		panic(err)
	}

	time.Sleep(400 * time.Millisecond)
	fmt.Printf("[ACTION][%s] rewrite dist/hi content\n", time.Now().Format("15:04:05.000"))
	if err := os.WriteFile(targetPath, []byte("v2: rewrite content\n"), 0644); err != nil {
		panic(err)
	}

	time.Sleep(500 * time.Millisecond)
	fmt.Printf("[ACTION][%s] replace dist/hi via rename\n", time.Now().Format("15:04:05.000"))
	if err := os.WriteFile(tempRenamePath, []byte("v3: content from rename replacement\n"), 0644); err != nil {
		panic(err)
	}
	if err := os.Rename(tempRenamePath, targetPath); err != nil {
		panic(err)
	}

	time.Sleep(1200 * time.Millisecond)
	mu.Lock()
	count := len(onChangeTimes)
	for i, ts := range onChangeTimes {
		fmt.Printf("[SUMMARY] onChange #%d at %s\n", i+1, ts.Format("15:04:05.000"))
	}
	mu.Unlock()
	fmt.Printf("[SUMMARY] total onChange calls = %d\n", count)
}
