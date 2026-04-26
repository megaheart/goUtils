package fsUtils

import (
	"bufio"
	"errors"
	"io"
	"os"
	"time"
)

type TailLine struct {
	Line string
	Err  error
}

type TailLines struct {
	Lines chan TailLine
	Close func() error
}

func TailFilePolling_DependOnINode(path string, skipOld bool, pollInterval time.Duration) (*TailLines, error) {
	f, err := os.Open(path) // keep current inode
	if err != nil {
		return nil, err
	}
	if skipOld {
		_, err = f.Seek(0, io.SeekEnd) // move cursor to end of file, only tail new lines
		if err != nil {
			f.Close()
			return nil, err
		}
	}
	lines := make(chan TailLine)
	isClose := false
	go func() {
		defer f.Close()
		defer close(lines)
		r := bufio.NewReader(f)
		for {
			if isClose {
				return
			}
			line, err := r.ReadString('\n')
			if err == nil {
				lines <- TailLine{Line: line, Err: nil}
				continue
			}
			if errors.Is(err, io.EOF) {
				if isClose {
					return
				}
				time.Sleep(pollInterval) // chờ log append thêm
				continue
			}
			lines <- TailLine{Line: "", Err: err}
			return
		}
	}()
	return &TailLines{
		Lines: lines,
		Close: func() error {
			isClose = true
			return f.Close()
		},
	}, nil

}
