package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// Explain all file operations that can be observed by the file watcher. These operations are represented as bitmask values,
// allowing for multiple operations to be combined.
//
// With FILE:
//   - Create: A new pathname was created.
//   - Write: The pathname was written to; this does *not* mean the write has finished, and a write can be followed by more writes.
//   - Remove: The path was removed; any watches on it will be removed. Some "remove" operations may trigger a Rename if the file is actually moved (for example "remove to trash" is often a rename).
//   - Rename: The path was renamed to something else; any watches on it will be removed.
//   - Chmod: File attributes were changed. It's generally not recommended to take action on this event, as it may get triggered very frequently by some software. For example, Spotlight indexing on macOS, anti-virus software, backup software, etc.
//   - xUnportableOpen: File descriptor was opened. Only works on Linux and FreeBSD.
//   - xUnportableRead: File was read from. Only works on Linux and FreeBSD.
//   - xUnportableCloseWrite: File opened for writing was closed. Only works on Linux and FreeBSD. The advantage of using this over Write is that it's more reliable than waiting for Write events to stop. It's also faster (if you're not listening to Write events): copying a file of a few GB can easily generate tens of thousands of Write events in a short span of time.
//   - xUnportableCloseRead: File opened for reading was closed. Only works on Linux and FreeBSD.
//
// With FOLDER:
//   - Create: A new child pathname was created in the folder.
//   - Write: A child pathname was written to in the folder; this does *not* mean the write has finished, and a write can be followed by more writes.
//   - Remove: A child path was removed from the folder; any watches on it will be removed. Some "remove" operations may trigger a Rename if the file is actually moved (for example "remove to trash" is often a rename).
//   - Rename: A child path was renamed to something else in the folder; any watches on it will be removed.
//   - Chmod: A child pathname's attributes were changed in the folder. It's generally not recommended to take action on this event, as it may get triggered very frequently by some software. For example, Spotlight indexing on macOS, anti-virus software, backup software, etc.
type FileOp uint32

// The operations watcher can trigger
const (
	// A new pathname was created.
	FileOp_Create FileOp = 1 << iota

	// The pathname was written to; this does *not* mean the write has finished,
	// and a write can be followed by more writes.
	FileOp_Write

	// The path was removed; any watches on it will be removed. Some "remove"
	// operations may trigger a Rename if the file is actually moved (for
	// example "remove to trash" is often a rename).
	FileOp_Remove

	// The path was renamed to something else; any watches on it will be
	// removed.
	FileOp_Rename

	// File attributes were changed.
	//
	// It's generally not recommended to take action on this event, as it may
	// get triggered very frequently by some software. For example, Spotlight
	// indexing on macOS, anti-virus software, backup software, etc.
	FileOp_Chmod

	// File descriptor was opened.
	//
	// Only works on Linux and FreeBSD.
	FileOp_xUnportableOpen

	// File was read from.
	//
	// Only works on Linux and FreeBSD.
	FileOp_xUnportableRead

	// File opened for writing was closed.
	//
	// Only works on Linux and FreeBSD.
	//
	// The advantage of using this over Write is that it's more reliable than
	// waiting for Write events to stop. It's also faster (if you're not
	// listening to Write events): copying a file of a few GB can easily
	// generate tens of thousands of Write events in a short span of time.
	FileOp_xUnportableCloseWrite

	// File opened for reading was closed.
	//
	// Only works on Linux and FreeBSD.
	FileOp_xUnportableCloseRead
)

func (op FileOp) String() string {
	var ops []string
	if op&FileOp_Create == FileOp_Create {
		ops = append(ops, "Create")
	}
	if op&FileOp_Write == FileOp_Write {
		ops = append(ops, "Write")
	}
	if op&FileOp_Remove == FileOp_Remove {
		ops = append(ops, "Remove")
	}
	if op&FileOp_Rename == FileOp_Rename {
		ops = append(ops, "Rename")
	}
	if op&FileOp_Chmod == FileOp_Chmod {
		ops = append(ops, "Chmod")
	}
	if op&FileOp_xUnportableOpen == FileOp_xUnportableOpen {
		ops = append(ops, "xUnportableOpen")
	}
	if op&FileOp_xUnportableRead == FileOp_xUnportableRead {
		ops = append(ops, "xUnportableRead")
	}
	if op&FileOp_xUnportableCloseWrite == FileOp_xUnportableCloseWrite {
		ops = append(ops, "xUnportableCloseWrite")
	}
	if op&FileOp_xUnportableCloseRead == FileOp_xUnportableCloseRead {
		ops = append(ops, "xUnportableCloseRead")
	}
	if len(ops) == 0 {
		return "Unknown"
	}
	return fmt.Sprintf("%v", ops)
}

const (
	// Exactly one of O_RDONLY, O_WRONLY, or O_RDWR must be specified.
	O_RDONLY int = syscall.O_RDONLY // open the file read-only.
	O_WRONLY int = syscall.O_WRONLY // open the file write-only.
	O_RDWR   int = syscall.O_RDWR   // open the file read-write.
	// The remaining values may be or'ed in to control behavior.
	O_APPEND int = syscall.O_APPEND // append data to the file when writing.
	O_CREATE int = syscall.O_CREAT  // create a new file if none exists.
	O_EXCL   int = syscall.O_EXCL   // used with O_CREATE, file must not exist.
	O_SYNC   int = syscall.O_SYNC   // open for synchronous I/O.
	O_TRUNC  int = syscall.O_TRUNC  // truncate regular writable file when opened.
)

var SkipDirErr = filepath.SkipDir
var SkipAllErr = filepath.SkipAll

// FileEvent represents an event occurring on a file within the filesystem.
// It encapsulates the path to the file and the corresponding file operation,
// allowing the handling of various filesystem changes such as creation, modification, or deletion.
type FileEvent struct {
	// The path to the file that triggered the event.
	Path string
	// The operation that was performed on the file.
	// This can include operations like Create, Write, Remove, Rename, etc.
	// The operations are represented as a bitmask, allowing for multiple operations
	// to be combined.
	// For example, if a file was both created and written to, the Op field
	// could contain both Create and Write operations.
	Op FileOp
}

// IFile is an interface that defines basic file operations.
// It abstracts file handling capabilities, providing a set of methods
// that must be implemented to interact with file-like objects.
// The interface includes methods for closing a file, reading from and writing to it,
// seeking to a specific position within the file, and obtaining its FileInfo metadata.
type IFile interface {
	io.Reader
	io.Writer
	// Close closes the file, releasing any resources associated with it.
	// It returns an error if the close operation fails.
	Close() error
	// Read reads data from the file into the provided byte slice.
	// It returns the number of bytes read and an error if the read operation fails.
	// The number of bytes read may be less than the length of the provided slice
	// if the end of the file is reached or if an error occurs.
	// If the read operation is successful, n will be the number of bytes read.
	Read(p []byte) (n int, err error)
	// Write writes data from the provided byte slice to the file.
	// It returns the number of bytes written and an error if the write operation fails.
	// The number of bytes written may be less than the length of the provided slice
	// if an error occurs during the write operation.
	// If the write operation is successful, n will be the number of bytes written.
	// If the file is opened in read-only mode, this method will return an error.
	Write(p []byte) (n int, err error)
	// Seek sets the offset for the next read or write operation on the file.
	// The offset is relative to the start of the file, and the whence parameter
	// specifies the reference point for the offset:
	//
	//   - 0 for the start of the file
	//   - 1 for the current position
	//   - 2 for the end of the file
	//
	// It returns the new offset and an error if the seek operation fails.
	// If the seek operation is successful, new offset returned will be the new position
	// of cursor in the file after the seek operation.
	// If the file is opened in read-only mode, this method will return an error.
	Seek(offset int64, whence int) (int64, error)
	// Stat returns the FileInfo structure describing the file.
	// The FileInfo structure contains metadata about the file, such as its name,
	// size, permissions, and modification time.
	// It returns the FileInfo and an error if the stat operation fails.
	Stat() (os.FileInfo, error)
	// Truncate truncates the file to the specified size.
	// It returns an error if the truncate operation fails.
	// If the file is opened in read-only mode, this method will return an error.
	Truncate(size int64) error
	// Name returns the name of the file.
	Name() string
	// WriteAt writes data to the file at the specified offset.
	// It returns the number of bytes written and an error if the write operation fails.
	WriteAt(p []byte, off int64) (n int, err error)
	// Sync synchronizes the file's in-memory state with the underlying storage.
	// It ensures that any changes made to the file are written to disk.
	// It returns an error if the sync operation fails.
	Sync() error
}

// IFileWatcher is an interface that defines methods for monitoring and managing file system paths.
// It allows clients to add or remove paths to be watched, retrieve the list of currently watched paths,
// and obtain channels for receiving file events and error notifications.
type IFileWatcher interface {
	// Add adds a new path to be watched for file system events.
	// It returns an error if the path cannot be added to the watch list.
	// The watcher will monitor the specified path for changes such as creation,
	// modification, or deletion of files and directories.
	Add(path string) error
	// Remove removes a path from the watch list.
	// It returns an error if the path cannot be removed from the watch list.
	// The watcher will no longer monitor the specified path for changes.
	Remove(path string) error
	// Close stops the watcher and releases any resources associated with it.
	// It returns an error if the close operation fails.
	// After closing, the watcher will no longer receive events or notifications
	Close() error
	// WatchedList returns a slice of strings representing the paths currently being watched.
	WatchedList() []string
	// Events returns a channel for receiving file system events.
	// The channel will deliver events such as file creation, modification,
	// deletion, and renaming.
	Events() <-chan FileEvent
	// Errors returns a channel for receiving error notifications.
	// The channel will deliver errors encountered while monitoring the file system.
	// This can include errors related to file access, permission issues,
	// or other unexpected conditions.
	Errors() <-chan error
}

// IFileSystem is an interface that defines a set of methods for file system operations.
// It provides an abstraction layer for file handling, allowing for operations
// such as reading, writing, and manipulating files and directories.
type IFileSystem interface {
	// ReadFile reads the contents of a file at the specified path.
	// It returns the file contents as a byte slice and an error if the read operation fails.
	ReadFile(path string) ([]byte, error)
	// WriteFileWithPerm writes data to a file at the specified path.
	// If the file does not exist, it will be created.
	// If the file already exists, it will be truncated before writing.
	// It returns an error if the write operation fails.
	WriteFileWithPerm(path string, data []byte, perm os.FileMode) error
	// AppendFile appends data to a file at the specified path.
	// If the file does not exist, it will be created.
	// It returns the number of bytes written and an error if the append operation fails.
	// If the file already exists, it will be opened in append mode.
	// If the file is opened in read-only mode, this method will return an error.
	AppendFile(path string, data []byte) (int, error)
	// Remove a file/folder at the specified path.
	// It returns an error if the remove operation fails.
	// If the file/folder does not exist, it will return an error. The error can be checked by
	// using the IsNotExistError method.
	Remove(path string) error
	// RemoveAll removes a file or directory at the specified path.
	// It returns an error if the remove operation fails.
	// If the file or directory does not exist, it will return an error. The error can be checked by
	// using the IsFileNotExistError or IsDirNotExistError methods.
	// If the path is a directory, it will be removed recursively.
	// If the path is a file, it will be removed.
	// If the path is a symbolic link, it will be removed.
	RemoveAll(path string) error
	// CleanDir removes all files and directories in the specified directory.
	// It does not remove the directory itself, only its contents.
	// It returns an error if the clean operation fails.
	CleanDir(dirPath string) error
	// MakeDir creates a directory at the specified path.
	// It returns an error if the create operation fails.
	// If the directory already exists, it will return an error. The error can be checked by
	// using the IsDirExistError method.
	// If the path is a file, it will return an error. The error can be checked by
	// using the IsFileExistError method.
	// If the path is a symbolic link, it will return an error. The error can be checked by
	// using the IsFileExistError method.
	MakeDir(path string, perm os.FileMode) error
	// MakeDirAll creates a directory and all necessary parent directories at the specified path.
	// It returns an error if the create operation fails.
	// If the directory already exists, it will return an error. The error can be checked by
	// using the IsDirExistError method.
	// If the path is a file, it will return an error. The error can be checked by
	// using the IsFileExistError method.
	// If the path is a symbolic link, it will return an error. The error can be checked by
	// using the IsFileExistError method.
	// If the path is a symbolic link, it will return an error. The error can be checked by
	// using the IsFileExistError method.
	MakeDirAll(path string, perm os.FileMode) error
	// MakeTempDir creates a temporary directory with a unique name whose prefix is specified.
	// It returns the path to the created temporary directory and an error if the create operation fails.
	// The temporary directory will be removed automatically when the program exits.
	MakeTempDir(dir, prefix string) (string, error)
	// MakeTempFile creates a temporary file with a unique name whose prefix is specified.
	// It returns the path to the created temporary file and an error if the create operation fails.
	// The temporary file will be removed automatically when the program exits.
	MakeTempFile(dir, prefix string) (string, error)
	// MakeTempFileWithContent creates a temporary file with the specified content and a unique name whose prefix is specified.
	// It returns the path to the created temporary file and an error if the create operation fails.
	// The temporary file will be removed automatically when the program exits.
	MakeTempFileWithContent(dir, prefix string, data []byte) (string, error)
	// MakeTempFileWithContentAndPerm creates a temporary file with the specified content, a unique name whose prefix is specified
	// and the specified permissions for the file.
	// It returns the path to the created temporary file and an error if the create operation fails.
	// The temporary file will be removed automatically when the program exits.
	MakeTempFileWithContentAndPerm(dir, prefix string, data []byte, perm os.FileMode) (string, error)
	// OpenFile opens a file at the specified path.
	// It returns the opened file and an error if the open operation fails.
	OpenFile(path string) (IFile, error)
	// OpenFileWithFlagsAndPerm opens a file at the specified path with the given flags and permissions.
	// The flags can include options such as read, write, append, etc.
	// The permissions specify the file mode (e.g., read, write, execute).
	// It returns the opened file and an error if the open operation fails.
	OpenFileWithFlagsAndPerm(path string, flags int, perm os.FileMode) (IFile, error)
	// CreateFile creates a new file at the specified path.
	// If the file already exists, it will be truncated before writing.
	// It returns the created file and an error if the create operation fails.
	CreateFile(path string) (IFile, error)
	// Stat returns the FileInfo structure describing the file at the specified path.
	// The FileInfo structure contains metadata about the file, such as its name,
	// size, permissions, and modification time.
	Stat(path string) (os.FileInfo, error)
	// IsFileExistError checks if the error is due to a file/folder already existing.
	// It returns true if the error indicates that the file/folder already exists,
	// and false otherwise.
	IsExistError(err error) bool
	// IsFileNotExistError checks if the error is due to a file/folder not existing.
	// It returns true if the error indicates that the file/folder does not exist,
	// and false otherwise.
	IsNotExistError(err error) bool
	// IsNotDirError checks if the error is due to a path not being a directory.
	// It returns true if the error indicates that the path is not a directory,
	// and false otherwise.
	IsNotDirError(err error) bool
	// IsNotFileError checks if the error is due to a path not being a file.
	// It returns true if the error indicates that the path is not a file,
	// and false otherwise.
	IsDirError(err error) bool
	// NewFileWatcher creates a new file watcher instance.
	// It returns an IFileWatcher interface and an error if the creation fails.
	// The file watcher can be used to monitor file system events such as
	// file creation, modification, deletion, and renaming.
	NewFileWatcher() (IFileWatcher, error)
	// CopyFile copies a file from the source path to the destination path.
	// It returns an error if the copy operation fails.
	CopyFile(src, dst string) error
	// CopyFileWithPerm copies a file from the source path to the destination path
	// and sets the specified permissions for the copied file.
	// It returns an error if the copy operation fails.
	CopyFileWithPerm(src, dst string, perm os.FileMode) error
	// MoveFile moves a file from the source path to the destination path.
	// It returns an error if the move operation fails.
	MoveFile(src, dst string) error
	// MoveFileWithPerm moves a file from the source path to the destination path
	// and sets the specified permissions for the moved file.
	// It returns an error if the move operation fails.
	MoveFileWithPerm(src, dst string, perm os.FileMode) error
	// GetFileAndFolderNamesInDir returns a list of files and folders'names in the specified directory.
	// It returns a slice of strings representing the file+folder names and an error if the operation fails.
	GetFileAndFolderNamesInDir(dirPath string) ([]string, error)
	// GetFileAndFolderInfosInDir returns a list of files and folders'infos in the specified directory.
	// It returns a slice of os.FileInfo representing the file+folder infos and an error if the operation fails.
	GetFileAndFolderInfosInDir(dirPath string) ([]os.FileInfo, error)
	// Walk walks the file tree rooted at root, calling walkFn for each file or directory in the tree.
	Walk(root string, walkFn func(path string, info os.FileInfo, err error) error) error
	// WalkDir walks the directory tree rooted at root, calling walkFn for each directory in the tree.
	WalkDir(root string, walkFn func(path string, d os.DirEntry, err error) error) error
	// Write from io.Reader to IFile
	WriteFromReader(file IFile, reader io.Reader) (int64, error)
	// Write from io.Reader to IFile with offset
	WriteFromReaderAt(file IFile, reader io.Reader, offset int64) (int64, error)
	// Create a symbolic link at newname pointing to oldname.
	// It returns an error if the symlink creation fails.
	Symlink(oldname, newname string) error
}
