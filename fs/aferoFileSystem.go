package fs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/spf13/afero"
	"golang.org/x/sys/unix"
)

// AferoFileWatcher implements IFileWatcher using fsnotify.
type AferoFileWatcher struct {
	eventsChan chan FileEvent
	errorsChan chan error
	watched    map[string]struct{}
	mu         sync.Mutex
}

// NewAferoFileWatcher returns a new AferoFileWatcher.
func NewAferoFileWatcher() *AferoFileWatcher {
	afw := &AferoFileWatcher{
		eventsChan: make(chan FileEvent),
		errorsChan: make(chan error),
		watched:    make(map[string]struct{}),
	}
	// go afw.dispatch()
	return afw
}

// Emit filw changes to the events channel.
// This function is called when a file change event is detected.
func (a *AferoFileWatcher) emit(path string, op FileOp) {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.watched[path]
	if !ok {
		return
	}

	event := FileEvent{
		Path: path,
		Op:   op,
	}

	a.eventsChan <- event
}

// Add registers the given path with the file watcher.
func (a *AferoFileWatcher) Add(path string) error {
	a.mu.Lock()
	a.watched[path] = struct{}{}
	a.mu.Unlock()
	return nil
}

// Remove unregisters the given path from the file watcher.
func (a *AferoFileWatcher) Remove(path string) error {
	a.mu.Lock()
	delete(a.watched, path)
	a.mu.Unlock()
	return nil
}

// Close stops watching all paths and closes the watcher.
func (a *AferoFileWatcher) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.watched = nil
	close(a.eventsChan)
	close(a.errorsChan)
	return nil
}

// WatchedList returns a slice of currently watched paths.
func (a *AferoFileWatcher) WatchedList() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	paths := make([]string, 0, len(a.watched))
	for p := range a.watched {
		paths = append(paths, p)
	}
	return paths
}

// Events returns the channel that delivers file events.
func (a *AferoFileWatcher) Events() <-chan FileEvent {
	return a.eventsChan
}

// Errors returns the channel that delivers errors.
func (a *AferoFileWatcher) Errors() <-chan error {
	return a.errorsChan
}

// AferoFileSystem implements IFileSystem using an afero.Fs.
type AferoFileSystem struct {
	fs       afero.Fs
	watchers []*AferoFileWatcher
}

// NewAferoFileSystem creates a new AferoFileSystem with the given afero.Fs.
func NewAferoFileSystem() *AferoFileSystem {
	var fs afero.Fs = afero.NewMemMapFs()
	return &AferoFileSystem{fs: fs}
}

// ReadFile reads the content of a file as a byte slice.
func (a *AferoFileSystem) ReadFile(path string) ([]byte, error) {
	return afero.ReadFile(a.fs, path)
}

// WriteFile writes data to a file with the given permissions.
func (a *AferoFileSystem) WriteFileWithPerm(path string, data []byte, perm os.FileMode) error {
	err := afero.WriteFile(a.fs, path, data, perm)

	if err == nil {
		for _, watcher := range a.watchers {
			if watcher != nil {
				watcher.emit(path, FileOp_Write)
			}
		}
	}

	return err
}

// AppendFile appends data to a file, creating it if necessary.
// Return the number of bytes written and any error encountered.
func (a *AferoFileSystem) AppendFile(path string, data []byte) (int, error) {
	f, err := a.fs.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var n int
	n, err = f.Write(data)

	if err == nil {
		for _, watcher := range a.watchers {
			if watcher != nil {
				watcher.emit(path, FileOp_Write)
			}
		}
	}
	return n, err
}

// Removes a file/folder at the given path
func (a *AferoFileSystem) Remove(path string) error {
	err := a.fs.Remove(path)

	if err == nil {
		for _, watcher := range a.watchers {
			if watcher != nil {
				watcher.emit(path, FileOp_Remove)
			}
		}
	}
	return err
}

// RemoveAll removes the directory and all its contents.
func (a *AferoFileSystem) RemoveAll(path string) error {
	err := a.fs.RemoveAll(path)

	if err == nil {
		for _, watcher := range a.watchers {
			if watcher != nil {
				watcher.emit(path, FileOp_Remove)
			}
		}
	}
	return err
}

// CleanDir removes all files and directories in the specified directory.
// It does not remove the directory itself, only its contents.
// It returns an error if the clean operation fails.
func (a *AferoFileSystem) CleanDir(dirPath string) error {
	// Read the directory contents
	infos, err := afero.ReadDir(a.fs, dirPath)
	if err != nil {
		return err
	}

	// Iterate over each item in the directory
	for _, info := range infos {
		itemPath := filepath.Join(dirPath, info.Name())
		if info.IsDir() {
			// Remove subdirectory and its contents
			if err := a.RemoveAll(itemPath); err != nil {
				return err
			}
		} else {
			// Remove file
			if err := a.Remove(itemPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// MakeDir creates a new directory with the specified permissions.
func (a *AferoFileSystem) MakeDir(path string, perm os.FileMode) error {
	// Check if the directory already exists
	fileInfo, err := a.fs.Stat(path)
	nonExist := true

	if err == nil {
		if fileInfo.IsDir() {
			nonExist = false
		}
	}

	err = a.fs.Mkdir(path, perm)

	if err == nil && nonExist {
		for _, watcher := range a.watchers {
			if watcher != nil {
				watcher.emit(path, FileOp_Create)
			}
		}
	}
	return err
}

// MakeDirAll creates a directory path and all necessary parents with the specified permissions.
func (a *AferoFileSystem) MakeDirAll(path string, perm os.FileMode) error {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(filepath.ToSlash(path), "/")
	if path == "" {
		return nil
	}
	path = "/" + path

	nonExistPaths := make([]string, 0)
	for {
		if path == "/" {
			break
		}
		file, err := a.fs.Stat(path)

		if err == nil {
			if !file.IsDir() {
				return &os.PathError{
					Op:   "mkdir",
					Path: path,
					Err:  unix.ENOTDIR,
				}
			}
			break
		} else if a.IsNotExistError(err) {
			nonExistPaths = append(nonExistPaths, path)
		} else {
			return err
		}

		path = filepath.Dir(path)
	}

	err := a.fs.MkdirAll(path, perm)

	if err == nil {
		for _, watcher := range a.watchers {
			if watcher != nil {
				for _, nonExistPath := range nonExistPaths {
					watcher.emit(nonExistPath, FileOp_Create)
				}
			}
		}
	}
	return err
}

// MakeTempDir creates a temporary directory in the specified directory with the given prefix.
func (a *AferoFileSystem) MakeTempDir(dir, prefix string) (string, error) {
	//NOTE: No sence when emit create event here
	return afero.TempDir(a.fs, dir, prefix)
}

// MakeTempFile creates a temporary file in the specified directory with the given prefix.
func (a *AferoFileSystem) MakeTempFile(dir, prefix string) (string, error) {
	//NOTE: No sence when emit create event here
	f, err := afero.TempFile(a.fs, dir, prefix)
	if err != nil {
		return "", err
	}
	name := f.Name()
	f.Close()
	return name, nil
}

// MakeTempFileWithContent creates a temporary file with the given initial data.
func (a *AferoFileSystem) MakeTempFileWithContent(dir, prefix string, data []byte) (string, error) {
	//NOTE: No sence when emit create event here
	f, err := afero.TempFile(a.fs, dir, prefix)
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		return "", err
	}
	f.Close()
	return name, nil
}

// MakeTempFileWithContentAndPerm creates a temporary file with the given data and sets its permissions.
func (a *AferoFileSystem) MakeTempFileWithContentAndPerm(dir, prefix string, data []byte, perm os.FileMode) (string, error) {
	//NOTE: No sence when emit create event here
	f, err := afero.TempFile(a.fs, dir, prefix)
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		return "", err
	}
	f.Close()
	if err := a.fs.Chmod(name, perm); err != nil {
		return "", err
	}
	return name, nil
}

// OpenFile opens a file and returns an *os.File.
// Note: This works correctly if the underlying fs is an OsFs.
func (a *AferoFileSystem) OpenFile(path string) (IFile, error) {
	// Fallback: open using afero and try to cast to *os.File.
	return a.fs.Open(path)
}

// OpenFileWithFlagsAndPerm opens a file with the specified flags and permissions.
func (a *AferoFileSystem) OpenFileWithFlagsAndPerm(path string, flags int, perm os.FileMode) (IFile, error) {
	// Fallback: open using afero and try to cast to *os.File.
	// Note: Afero does not support flags like O_APPEND, O_CREATE, etc. directly.
	// It uses os.O_RDWR by default, so we can only use it for reading/writing.
	return a.fs.OpenFile(path, flags, perm)
}

// CreateFile creates a file at the specified path with the given permissions.
func (a *AferoFileSystem) CreateFile(path string) (IFile, error) {
	// Fallback: create using afero and try to cast to *os.File.
	return a.fs.Create(path)
}

// Stat retrieves the file information for the given path.
func (a *AferoFileSystem) Stat(path string) (os.FileInfo, error) {
	return a.fs.Stat(path)
}

// IsDirExistError checks if the error is a directory exists error.
func (fs *AferoFileSystem) IsExistError(err error) bool {
	return os.IsExist(err)
}

// IsDirNotExistError checks if the error is a directory not exists error.
func (fs *AferoFileSystem) IsNotExistError(err error) bool {
	return os.IsNotExist(err)
}

// IsFileExistError checks if the error is a file exists error.
func (fs *AferoFileSystem) IsDirError(err error) bool {
	panic("IsDirError is not implemented")
}

// IsFileExistError checks if the error is a file exists error.
func (fs *AferoFileSystem) IsNotDirError(err error) bool {
	panic("IsNotDirError is not implemented")
}

// NewFileWatcher creates a new AferoFileWatcher.
func (a *AferoFileSystem) NewFileWatcher() (IFileWatcher, error) {
	w := NewAferoFileWatcher()
	a.watchers = append(a.watchers, w)
	return w, nil
}

// CopyFile copies a file from the source path to the destination path.
// It returns an error if the copy operation fails.
func (a *AferoFileSystem) CopyFile(src, dst string) error {
	srcFile, err := a.fs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := a.fs.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CopyFileWithPerm copies a file from the source path to the destination path
// and sets the specified permissions for the copied file.
// It returns an error if the copy operation fails.
func (a *AferoFileSystem) CopyFileWithPerm(src, dst string, perm os.FileMode) error {
	if err := a.CopyFile(src, dst); err != nil {
		return err
	}
	return a.fs.Chmod(dst, perm)
}

// MoveFile moves a file from the source path to the destination path.
// It returns an error if the move operation fails.
func (a *AferoFileSystem) MoveFile(src, dst string) error {
	return a.fs.Rename(src, dst)
}

// MoveFileWithPerm moves a file from the source path to the destination path
// and sets the specified permissions for the moved file.
// It returns an error if the move operation fails.
func (a *AferoFileSystem) MoveFileWithPerm(src, dst string, perm os.FileMode) error {
	if err := a.fs.Rename(src, dst); err != nil {
		return err
	}
	return a.fs.Chmod(dst, perm)
}

// GetFileAndFolderNamesInDir returns a list of files and folders'names in the specified directory.
// It returns a slice of strings representing the file+folder names and an error if the operation fails.
func (a *AferoFileSystem) GetFileAndFolderNamesInDir(dirPath string) ([]string, error) {
	infos, err := afero.ReadDir(a.fs, dirPath)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}
	return names, nil
}

// GetFileAndFolderInfosInDir returns a list of files and folders'infos in the specified directory.
// It returns a slice of os.FileInfo representing the file+folder infos and an error if the operation fails.
func (a *AferoFileSystem) GetFileAndFolderInfosInDir(dirPath string) ([]os.FileInfo, error) {
	infos, err := afero.ReadDir(a.fs, dirPath)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

// walkDir recursively descends path, calling walkDirFn.
func (a *AferoFileSystem) walkDir(path string, d os.DirEntry, walkDirFn func(path string, d os.DirEntry, err error) error) error {
	if err := walkDirFn(path, d, nil); err != nil || !d.IsDir() {
		if err == filepath.SkipDir && d.IsDir() {
			// Successfully skipped directory.
			err = nil
		}
		return err
	}

	dirs, err := afero.ReadDir(a.fs, path)
	if err != nil {
		// Second call, to report ReadDir error.
		err = walkDirFn(path, d, err)
		if err != nil {
			if err == filepath.SkipDir && d.IsDir() {
				err = nil
			}
			return err
		}
	}

	for _, d1 := range dirs {
		path1 := filepath.Join(path, d1.Name())
		if err := a.walkDir(path1, fs.FileInfoToDirEntry(d1), walkDirFn); err != nil {
			if err == filepath.SkipDir {
				break
			}
			return err
		}
	}
	return nil
}

// walk recursively descends path, calling walkFn.
func (a *AferoFileSystem) walk(path string, info os.FileInfo, walkFn func(path string, info os.FileInfo, err error) error) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	f, err := a.fs.Open(path)
	if err != nil {
		return walkFn(path, info, err)
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return walkFn(path, info, err)
	}
	slices.Sort(names)

	err1 := walkFn(path, info, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means walkFn want walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by walkFn. walkFn may ignore err and return nil.
		// If walkFn returns SkipDir or SkipAll, it will be handled by the caller.
		// So walk should return whatever walkFn returns.
		return err1
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := a.fs.Stat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = a.walk(filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

// Walk walks the file tree rooted at root, calling walkFn for each file or directory in the tree.
func (a *AferoFileSystem) Walk(root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	info, err := a.fs.Stat(root)
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = a.walk(root, info, walkFn)
	}
	if err == filepath.SkipDir || err == filepath.SkipAll {
		return nil
	}
	return err
}

// WalkDir walks the directory tree rooted at root, calling walkFn for each directory in the tree.
func (a *AferoFileSystem) WalkDir(root string, walkFn func(path string, d os.DirEntry, err error) error) error {
	info, err := a.fs.Stat(root)
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = a.walkDir(root, fs.FileInfoToDirEntry(info), walkFn)
	}
	if err == filepath.SkipDir || err == filepath.SkipAll {
		return nil
	}
	return err
}

// Write from io.Reader to IFile
func (fs *AferoFileSystem) WriteFromReader(file IFile, reader io.Reader) (int64, error) {
	return io.Copy(file, reader)
}

// Write from io.Reader to IFile with offset
func (fs *AferoFileSystem) WriteFromReaderAt(file IFile, reader io.Reader, offset int64) (int64, error) {
	if code, err := file.Seek(int64(offset), io.SeekStart); err != nil {
		return code, err
	}

	return io.Copy(file, reader)
}
