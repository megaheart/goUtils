package fsUtils

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/megaheart/goUtils/fs"
	"github.com/megaheart/goUtils/log"
)

type FileWatchMode int

const (
	// FileWatchMode_Modify: The onChange function will be called when the file is modified (e.g. write to the file, inode is replaced)
	FileWatchMode_Modify FileWatchMode = iota
	// FileWatchMode_Replace: The onChange function will be called when the file is replaced with new inode (e.g. delete the file and create a new file with the same name)
	FileWatchMode_Replace
	// FileWatchMode_Remove: The onChange function will be called when the file is removed (e.g. delete the file, move the file, rename the file)
	FileWatchMode_Remove
)

func (m FileWatchMode) String() string {
	switch m {
	case FileWatchMode_Modify:
		return "Modify"
	case FileWatchMode_Replace:
		return "Replace"
	case FileWatchMode_Remove:
		return "Remove"
	default:
		return "Unknown"
	}
}

type FileWatchInfo struct {
	OnChanged func() error
	Mode      FileWatchMode
	// inactive   bool
	ActiveTime time.Time
}

type WatchFileConflictError struct {
	Path string
}

func (e *WatchFileConflictError) Error() string {
	return "Conflict watch file: " + e.Path
}

type WatchFolderAsFileError struct {
	Path string
}

func (e *WatchFolderAsFileError) Error() string {
	return "Can't watch folder as file: " + e.Path
}

// NOT THREAD-SAFE
type FileWatcher struct {
	logger            log.ILogger
	watcher           fs.IFileWatcher
	watchFileEventMap map[string][]*FileWatchInfo
	watchFolderMap    map[string]int
	fs                fs.IFileSystem
	DebounceTime      time.Duration
}

// NOT THREAD-SAFE
//
// Watch a target event of a specified file, supported event:
//
//   - Modify: The onChange function will be called when the file is modified (e.g. write to the file, inode is replaced)
//   - Replace: The onChange function will be called when the file is replaced with new inode (e.g. delete the file and create a new file with the same name)
//   - Remove: The onChange function will be called when the file is removed (e.g. delete the file, move the file, rename the file)
func NewFileWatcher(
	logger log.ILogger,
	fs fs.IFileSystem,
	debounceTime time.Duration,
) *FileWatcher {
	return &FileWatcher{
		watcher:           nil,
		watchFileEventMap: make(map[string][]*FileWatchInfo),
		logger:            logger,
		fs:                fs,
		DebounceTime:      debounceTime,
	}
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
func (fw *FileWatcher) WatchFile(path string, mode FileWatchMode, onChange func() error) error {
	// Only allow to call this function when not in system runtime (call on main routine)
	// becase FileWatcher is not thread-safe
	// fw.systemCtrl.BlockRuntimeCall("FileWatcher.WatchFile")
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		fw.logger.Fatal("Error getting absolute path: " + err.Error())
	}

	stat, err := fw.fs.Stat(absolutePath)
	if err != nil {
		if !fw.fs.IsNotExistError(err) {
			return err
		}
	} else if stat.IsDir() {
		return &WatchFolderAsFileError{Path: absolutePath}
	}

	err = onChange()

	if err != nil {
		fw.logger.Fatal("Error calling onChange function when init watcher: " + err.Error())
	}

	if fw.watcher == nil {
		fw.InitWatcher()
	}

	if err != nil {
		fw.logger.Error("Error while adding file to watcher", log.LogError(err))
	}

	if f, ok := fw.watchFileEventMap[absolutePath]; ok {
		f = append(f, &FileWatchInfo{
			OnChanged:  onChange,
			ActiveTime: time.Time{},
			Mode:       mode,
			// inactive:   true,
		})
		return nil
	}

	fw.watchFileEventMap[absolutePath] = []*FileWatchInfo{
		{
			OnChanged:  onChange,
			ActiveTime: time.Time{},
			Mode:       mode,
			// inactive:   true,
		},
	}

	parentDir := filepath.Dir(absolutePath)
	if parentDir == "" {
		return errors.New("Failed to get parent directory of " + absolutePath)
	}

	if num, ok := fw.watchFolderMap[parentDir]; ok {
		fw.watchFolderMap[parentDir] = num + 1
	} else {
		fw.watchFolderMap[parentDir] = 1
		err = fw.watcher.Add(parentDir)
		if err != nil {
			fw.logger.Error("Error while adding file to watcher", log.LogError(err))
		} else {
			fw.logger.Info("Watching file: " + absolutePath + " with mode: " + mode.String())
		}
		return err
	}

	return nil
}

const fileOps_FOR_FILEWATCHMODE_REPLACE = fs.FileOp_Create
const fileOps_FOR_FILEWATCHMODE_MODIFY = fs.FileOp_Write | fs.FileOp_xUnportableCloseWrite | fs.FileOp_Create | fileOps_FOR_FILEWATCHMODE_REPLACE
const fileOps_FOR_FILEWATCHMODE_REMOVE = fs.FileOp_Remove | fs.FileOp_Rename

func (fw *FileWatcher) eventOccurred(event fs.FileEvent) {
	path := event.Path
	if pathInfos, ok := fw.watchFileEventMap[path]; ok {
		for _, f := range pathInfos {
			switch f.Mode {
			case FileWatchMode_Modify:
				if event.Op&fs.FileOp_Write != 0 {
					if f.ActiveTime.After(time.Now()) {
						continue
					}
					f.ActiveTime = time.Now().Add(fw.DebounceTime)
					go func() {
						time.Sleep(fw.DebounceTime)
						err := f.OnChanged()
						if err == nil {
							fw.logger.Info("OnChange function called successfully with file: " + path)
						} else {
							fw.logger.Error("Error calling onChange function with file: "+path, log.LogError(err))
						}
					}()
				} else if event.Op&fileOps_FOR_FILEWATCHMODE_MODIFY != 0 {
					go func() {
						err := f.OnChanged()
						if err == nil {
							fw.logger.Info("OnChange function called successfully with file: " + path)
						} else {
							fw.logger.Error("Error calling onChange function with file: "+path, log.LogError(err))
						}
					}()
				}

			case FileWatchMode_Replace:
				if event.Op&fileOps_FOR_FILEWATCHMODE_REPLACE != 0 {
					go func() {
						err := f.OnChanged()
						if err == nil {
							fw.logger.Info("OnChange function called successfully with file: " + path)
						} else {
							fw.logger.Error("Error calling onChange function with file: "+path, log.LogError(err))
						}
					}()
				}
			case FileWatchMode_Remove:
				if event.Op&fileOps_FOR_FILEWATCHMODE_REMOVE != 0 {
					go func() {
						err := f.OnChanged()
						if err == nil {
							fw.logger.Info("OnChange function called successfully with file: " + path)
						} else {
							fw.logger.Error("Error calling onChange function with file: "+path, log.LogError(err))
						}
					}()
				}
			default:
				fw.logger.Error("Unsupported file watch mode: " + f.Mode.String())
			}
		}
	}
}

func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case event, ok := <-fw.watcher.Events():
			if !ok {
				fw.logger.Error("Watcher stop watching file because fileWatcher.Events channel is closed (ok = false, panic call)")
				return
			}
			fw.eventOccurred(event)

		case err, ok := <-fw.watcher.Errors():
			if !ok {
				if err == nil {
					err = errors.New("error=nil, watcher stop watching file because fileWatcher.Errors channel is closed")
				}
				fw.logger.Error("Watcher stop watching file because fileWatcher.Errors channel is closed", log.LogError(err))
				// system.Fatal("Error watching file (ok = false, panic call)", 1)
				return
			}
			if err == nil {
				err = errors.New("error=nil, ok=true, the behavior of the watcher.Errors function is unknown")
			} else {
				fw.logger.Error("Error watching file", log.LogError(err))
				// return
			}
		}
	}
}

func (fw *FileWatcher) Close() error {
	if fw.watcher != nil {
		return fw.watcher.Close()
	}
	return nil
}

// Initializa watcher
func (fw *FileWatcher) InitWatcher() {
	var err error
	fw.watcher, err = fw.fs.NewFileWatcher()

	if err != nil {
		fw.logger.Error("Error creating watcher", log.LogError(err))
	}

	go func() {
		for {
			fw.watchLoop()
			fw.watcher.Close()

			fw.logger.Info("Restarting watcher")
			fw.watcher, err = fw.fs.NewFileWatcher()
			if err != nil {
				fw.logger.Error("Error creating watcher", log.LogError(err))
			}

			for path := range fw.watchFileEventMap {
				err = fw.watcher.Add(path)
				if err != nil {
					fw.logger.Error("Error while adding file to watcher", log.LogError(err))
				}
			}
		}
	}()
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

	fw.WatchFile(path, FileWatchMode_Modify, f)
}

type WatchJsonFileFunc = func(string, func(interface{}))
