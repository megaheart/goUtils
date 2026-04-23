package fs

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/sys/unix"
)

type OsFileWatcher struct {
	watcher *fsnotify.Watcher
	events  chan FileEvent
	one     sync.Once
}

func (w *OsFileWatcher) Add(path string) error {
	return w.watcher.Add(path)
}
func (w *OsFileWatcher) Remove(path string) error {
	return w.watcher.Remove(path)
}
func (w *OsFileWatcher) Close() error {
	return w.watcher.Close()
}
func (w *OsFileWatcher) WatchedList() []string {
	return w.watcher.WatchList()
}
func (w *OsFileWatcher) Events() <-chan FileEvent {
	w.one.Do(func() {
		go func() {
			for evt := range w.watcher.Events {
				w.events <- FileEvent{
					Path: evt.Name,
					Op:   FileOp(evt.Op),
				}
			}
			close(w.events)
		}()
	})
	return w.events
}
func (w *OsFileWatcher) Errors() <-chan error {
	return w.watcher.Errors
}

type OsFileSystem struct {
}

// NewFileSystem creates a new instance of FileSystem.
func NewFileSystem() *OsFileSystem {
	return &OsFileSystem{}
}

// ReadFile reads a file from the given path.
func (fs *OsFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFileWithPerm writes data to a file with the specified permissions.
func (fs *OsFileSystem) WriteFileWithPerm(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// AppendFile appends data to the file at the given path.
// Return the number of bytes written and any error encountered.
func (fs *OsFileSystem) AppendFile(path string, data []byte) (int, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Write(data)
}

// Removes a file/folder at the given path.
func (fs *OsFileSystem) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll removes any file or directory at the given path.
func (fs *OsFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// CleanDir removes all files and directories in the specified directory.
// It does not remove the directory itself, only its contents.
// It returns an error if the clean operation fails.
func (fs *OsFileSystem) CleanDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		} else {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}

// MakeDir creates a new directory with the specified permissions.
func (fs *OsFileSystem) MakeDir(path string, perm os.FileMode) error {
	return os.Mkdir(path, perm)
}

// MakeDirAll creates a directory and all necessary parents.
func (fs *OsFileSystem) MakeDirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// MakeTempDir creates a temporary directory.
func (fs *OsFileSystem) MakeTempDir(dir, prefix string) (string, error) {
	return os.MkdirTemp(dir, prefix)
}

// MakeTempFile creates a temporary file and closes it immediately.
func (fs *OsFileSystem) MakeTempFile(dir, prefix string) (string, error) {
	f, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return "", err
	}
	name := f.Name()
	f.Close()
	return name, nil
}

// MakeTempFileWithContent creates a temporary file, writes data to it, and then closes it.
func (fs *OsFileSystem) MakeTempFileWithContent(dir, prefix string, data []byte) (string, error) {
	f, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err = f.Write(data); err != nil {
		f.Close()
		return "", err
	}
	return name, f.Close()
}

// MakeTempFileWithContentAndPerm creates a temporary file, writes data to it, sets the file permission, and then closes it.
func (fs *OsFileSystem) MakeTempFileWithContentAndPerm(dir, prefix string, data []byte, perm os.FileMode) (string, error) {
	f, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err = f.Write(data); err != nil {
		f.Close()
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = os.Chmod(name, perm); err != nil {
		return "", err
	}
	return name, nil
}

func underlyingErrorIs(err, target error) bool {
	// Note that this function is not errors.Is:
	// underlyingError only unwraps the specific error-wrapping types
	// that it historically did, not all errors implementing Unwrap().
	err = underlyingError(err)
	if err == target {
		return true
	}
	// To preserve prior behavior, only examine syscall errors.
	e, ok := err.(syscall.Errno)
	return ok && e.Is(target)
}

// underlyingError returns the underlying error for known os error types.
func underlyingError(err error) error {
	switch err := err.(type) {
	case *os.PathError:
		return err.Err
	case *os.LinkError:
		return err.Err
	case *os.SyscallError:
		return err.Err
	}
	return err
}

// OpenFile opens the file at the given path.
func (fs *OsFileSystem) OpenFile(path string) (IFile, error) {
	return os.Open(path)
}

// OpenFileWithFlagsAndPerm opens a file with the specified flags and permissions.
func (fs *OsFileSystem) OpenFileWithFlagsAndPerm(path string, flags int, perm os.FileMode) (IFile, error) {
	// Fallback: open using afero and try to cast to *os.File.
	// Note: Afero does not support flags like O_APPEND, O_CREATE, etc. directly.
	// It uses os.O_RDWR by default, so we can only use it for reading/writing.
	file, err := os.OpenFile(path, flags, perm)
	return file, err
}

// CreateFile creates a file at the specified path with the given permissions.
func (fs *OsFileSystem) CreateFile(path string) (IFile, error) {
	return os.Create(path)
}

// Stat retrieves the file information for the given path.
func (fs *OsFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// IsDirExistError checks if the error is a directory exists error.
func (fs *OsFileSystem) IsExistError(err error) bool {
	return os.IsExist(err)
}

// IsDirNotExistError checks if the error is a directory not exists error.
func (fs *OsFileSystem) IsNotExistError(err error) bool {
	return os.IsNotExist(err)
}

// IsFileExistError checks if the error is a file exists error.
func (fs *OsFileSystem) IsDirError(err error) bool {
	return underlyingErrorIs(err, unix.EISDIR)
}

// IsFileExistError checks if the error is a file exists error.
func (fs *OsFileSystem) IsNotDirError(err error) bool {
	return underlyingErrorIs(err, unix.ENOTDIR)
}

// Return new FileWatcher instance to watch changes in files and directories.
func (fs *OsFileSystem) NewFileWatcher() (IFileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &OsFileWatcher{
		watcher: watcher,
		events:  make(chan FileEvent),
	}, nil
}

// CopyFile copies a file from src to dst.
// It creates a new file if dst does not exist, or overwrites it if it does.
// The file permissions are preserved.
// If the source file does not exist or is not a regular file, an error is returned.
// If the destination file cannot be created or written to, an error is returned.
func (fs *OsFileSystem) CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return &os.PathError{Op: "copy", Path: src, Err: os.ErrInvalid}
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	err = destination.Sync()
	if err != nil {
		return err
	}

	return nil
}

// CopyFileWithPerm copies a file from src to dst with the specified permissions.
// It creates a new file if dst does not exist, or overwrites it if it does.
// The file permissions are preserved.
// If the source file does not exist or is not a regular file, an error is returned.
// If the destination file cannot be created or written to, an error is returned.
func (fs *OsFileSystem) CopyFileWithPerm(src, dst string, perm os.FileMode) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return &os.PathError{Op: "copy", Path: src, Err: os.ErrInvalid}
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	err = destination.Sync()
	if err != nil {
		return err
	}

	return nil
}

// MoveFile moves a file from src to dst.
// It creates a new file if dst does not exist, or overwrites it if it does.
// The file permissions are preserved.
// If the source file does not exist or is not a regular file, an error is returned.
// If the destination file cannot be created or written to, an error is returned.
func (fs *OsFileSystem) MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// MoveFileWithPerm moves a file from src to dst with the specified permissions.
// It creates a new file if dst does not exist, or overwrites it if it does.
// The file permissions are preserved.
// If the source file does not exist or is not a regular file, an error is returned.
// If the destination file cannot be created or written to, an error is returned.
func (fs *OsFileSystem) MoveFileWithPerm(src, dst string, perm os.FileMode) error {
	if err := os.Rename(src, dst); err != nil {
		return err
	}
	return os.Chmod(dst, perm)
}

// GetFileAndFolderNamesInDir returns a list of files and folders'name in the specified directory.
// It returns an error if the directory does not exist or cannot be read.
func (fs *OsFileSystem) GetFileAndFolderNamesInDir(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string = make([]string, 0, len(entries))

	for _, entry := range entries {
		// if entry.Name() == "" {
		// 	continue
		// }
		files = append(files, entry.Name())
	}

	return files, nil
}

// GetFileAndFolderInfosInDir returns a list of files and folders'infos in the specified directory.
// It returns a slice of os.FileInfo representing the file+folder infos and an error if the operation fails.
func (fs *OsFileSystem) GetFileAndFolderInfosInDir(dirPath string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []os.FileInfo = make([]os.FileInfo, 0, len(entries))

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		// if info.Name() == "" {
		// 	continue
		// }
		files = append(files, info)
	}

	return files, nil
}

// Walk walks the file tree rooted at root, calling walkFn for each file or directory in the tree.
func (fs *OsFileSystem) Walk(root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	return filepath.Walk(root, walkFn)
}

// WalkDir walks the directory tree rooted at root, calling walkFn for each directory in the tree.
func (fs *OsFileSystem) WalkDir(root string, walkFn func(path string, d os.DirEntry, err error) error) error {
	return filepath.WalkDir(root, walkFn)
}

// Write from io.Reader to IFile
func (fs *OsFileSystem) WriteFromReader(file IFile, reader io.Reader) (int64, error) {
	return io.Copy(file, reader)
}

// Write from io.Reader to IFile with offset
func (fs *OsFileSystem) WriteFromReaderAt(file IFile, reader io.Reader, offset int64) (int64, error) {
	if _, err := file.Seek(int64(offset), io.SeekStart); err != nil {
		return 0, err
	}

	return io.Copy(file, reader)
}

// Symlink creates a symbolic link at newname pointing to oldname.
// It returns an error if the symlink creation fails.
func (fs *OsFileSystem) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}
