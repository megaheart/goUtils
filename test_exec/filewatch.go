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
	"github.com/nxadm/tail"
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
	logPath, err := filepath.Abs("dist/log.log")
	if err != nil {
		panic(err)
	}
	rotatedPath := logPath + ".1"

	_ = os.Remove(rotatedPath)
	if err := os.WriteFile(logPath, []byte("[old-1] seed old log\n"), 0644); err != nil {
		panic(err)
	}

	t, err := tail.TailFile(logPath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Poll:      true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: 0,
		},
	})
	if err != nil {
		panic(err)
	}
	defer t.Cleanup()
	defer t.Stop()

	go func() {
		for line := range t.Lines {
			if line == nil {
				continue
			}
			fmt.Printf("[TAIL][%s] %s\n", time.Now().Format("15:04:05.000"), line.Text)
		}
	}()

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

		data, readErr := os.ReadFile(logPath)
		if readErr != nil {
			fmt.Printf("[onChange #%d][%s] read error: %v\n", idx, now.Format("15:04:05.000"), readErr)
			return nil
		}
		fmt.Printf("[onChange #%d][%s] current dist/log.log content:\n%s", idx, now.Format("15:04:05.000"), string(data))
		return nil
	}

	if err := fw.WatchFile(logPath, fsUtils.FileWatchMode_Replace, onChange); err != nil {
		panic(err)
	}

	time.Sleep(400 * time.Millisecond)
	if err := os.WriteFile(logPath, []byte("[old-1] seed old log\n[old-2] append before rotate\n"), 0644); err != nil {
		panic(err)
	}

	time.Sleep(400 * time.Millisecond)
	rotateStart := time.Now()
	fmt.Printf("[ACTION][%s] start rotate dist/log.log\n", rotateStart.Format("15:04:05.000"))
	if err := os.Rename(logPath, rotatedPath); err != nil {
		panic(err)
	}
	if err := os.WriteFile(logPath, []byte("[new-1] new file after rotate\n"), 0644); err != nil {
		panic(err)
	}
	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		_, _ = f.WriteString("[new-2] append after rotate\n")
		_ = f.Close()
	}

	time.Sleep(1200 * time.Millisecond)
	mu.Lock()
	count := len(onChangeTimes)
	for i, ts := range onChangeTimes {
		fmt.Printf("[SUMMARY] onChange #%d at %s (delta from rotate start: %s)\n", i+1, ts.Format("15:04:05.000"), ts.Sub(rotateStart).String())
	}
	mu.Unlock()
	fmt.Printf("[SUMMARY] total onChange calls = %d\n", count)
	fmt.Printf("[SUMMARY] rotated file kept at: %s\n", rotatedPath)
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
