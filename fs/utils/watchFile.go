package fsUtils

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/megaheart/goUtils/fs"
	"github.com/megaheart/goUtils/log"
)

type FileWatchInfo struct {
	OnChanged  func() error
	ActiveTime time.Time
}

// NOT THREAD-SAFE
type FileWatcher struct {
	logger            log.ILogger
	watcher           fs.IFileWatcher
	watchFileEventMap map[string]*FileWatchInfo
	fs                fs.IFileSystem
	DebounceTime      time.Duration
}

// NOT THREAD-SAFE
func NewFileWatcher(
	logger log.ILogger,
	fs fs.IFileSystem,
	debounceTime time.Duration,
) *FileWatcher {
	return &FileWatcher{
		watcher:           nil,
		watchFileEventMap: make(map[string]*FileWatchInfo),
		logger:            logger,
		fs:                fs,
		DebounceTime:      debounceTime,
	}
}

// Initializa watcher
func (fw *FileWatcher) InitWatcher() {
	logger := fw.logger

	var err error
	fw.watcher, err = fw.fs.NewFileWatcher()

	if err != nil {
		logger.Error("Error creating watcher", log.LogError(err))
	}

	go func() {
		for {
			exitChan := make(chan error, 1)

			go (func() {
				for {
					select {
					case event, ok := <-fw.watcher.Events():
						if !ok {
							return
						}
						if event.Op&fs.FileOp_Write == fs.FileOp_Write {
							logger.Info("File modified: " + event.Path)
							if f, ok := fw.watchFileEventMap[event.Path]; ok {
								if f.ActiveTime.After(time.Now()) {
									continue
								}
								f.ActiveTime = time.Now().Add(fw.DebounceTime)
								go func() {
									time.Sleep(fw.DebounceTime)
									err := f.OnChanged()
									if err == nil {
										logger.Info("OnChange function called successfully with file: " + event.Path)
									} else {
										logger.Error("Error calling onChange function with file: "+event.Path, log.LogError(err))
									}
								}()
							}
						}
					case err, ok := <-fw.watcher.Errors():
						if !ok {
							if err == nil {
								err = errors.New("error=nil, watcher stop watching file because of being not ok")
							}
							logger.Error("Error watching file (ok = false, panic call)", log.LogError(err))
							// system.Fatal("Error watching file (ok = false, panic call)", 1)
							exitChan <- err
							panic(err)
						}
						if err == nil {
							err = errors.New("error=nil, the behavior of the watcher.Errors function is unknown")
						}
						logger.Error("Error watching file", log.LogError(err))
					}
				}
			})()

			<-exitChan
			fw.watcher.Close()

			logger.Info("Restarting watcher")
			fw.watcher, err = fw.fs.NewFileWatcher()
			if err != nil {
				logger.Error("Error creating watcher", log.LogError(err))
			}

			for path, _ := range fw.watchFileEventMap {
				err = fw.watcher.Add(path)
				if err != nil {
					logger.Error("Error while adding file to watcher", log.LogError(err))
				}
			}
		}
	}()
}

// The onChange function will be called when WatchJsonFile is called and the file is modified
//
// Parameters:
//
//   - path: The path to the JSON file
//
//   - onChange: The function to be called when WatchJsonFile is called and the file is modified
//
// Example:
//
//	WatchFile("foo.json", func() {\\ Do something})
func (fw *FileWatcher) WatchFile(path string, onChange func() error) {
	// Only allow to call this function when not in system runtime (call on main routine)
	// becase FileWatcher is not thread-safe
	// fw.systemCtrl.BlockRuntimeCall("FileWatcher.WatchFile")
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		fw.logger.Fatal("Error getting absolute path: " + err.Error())
	}

	err = onChange()

	if err != nil {
		fw.logger.Fatal("Error calling onChange function when init watcher: " + err.Error())
	}

	if fw.watcher == nil {
		fw.InitWatcher()
	}

	err = fw.watcher.Add(absolutePath)
	if err != nil {
		fw.logger.Error("Error while adding file to watcher", log.LogError(err))
	}

	fw.watchFileEventMap[absolutePath] = &FileWatchInfo{
		OnChanged:  onChange,
		ActiveTime: time.Time{},
	}
	fw.logger.Info("Watching file: " + absolutePath)
}

func (fw *FileWatcher) Close() error {
	if fw.watcher != nil {
		return fw.watcher.Close()
	}
	return nil
}

// Reads a JSON file and parses it into a struct
//
// Parameters:
//
//   - path: The path to the JSON file
//
// Returns:
//
//   - T: The struct that the JSON file is parsed into
//
//   - error: The error that occurred while reading the file
//
// Example:
//
//	foo, err := ReadJsonFile[[]Foo]("foo.json")
func ReadJsonFile[T any](fs fs.IFileSystem, path string) (T, error) {
	// Parse JSON thành slice của struct
	var obj T

	// Read all data from the file
	data, err := fs.ReadFile(path)
	if err != nil {
		return obj, err
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		// log.Fatal(err)
		return obj, err
	}

	return obj, nil
}

// The onChange function will be called when WatchJsonFile is called and the file is modified
//
// Parameters:
//
//   - path: The path to the JSON file
//
//   - onChange: The function to be called when WatchJsonFile is called and the file is modified
//
// Returns:
//
//   - *fsnotify.Watcher: The watcher that watches the file
//
// Example:
//
//	WatchJsonFile("foo.json", func(data []Foo) {\\ Do something})
func WatchJsonFile[T any](fw *FileWatcher, path string, onChange func(obj T) error) {
	f := func() error {
		obj, err := ReadJsonFile[T](fw.fs, path)
		if err != nil {
			fw.logger.Error("Error reading JSON file", log.LogError(err))
			return err
		}

		return onChange(obj)
	}

	fw.WatchFile(path, f)
}

type WatchJsonFileFunc = func(string, func(interface{}))
