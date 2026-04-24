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

func TailFilePolling_DependOnINode(path string, skipOld bool, pollInterval, timeout time.Duration) (chan TailLine, error) {
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
	go func() {
		defer f.Close()
		defer close(lines)
		r := bufio.NewReader(f)
		latestTailTime := time.Now()
		for {
			line, err := r.ReadString('\n')
			if err == nil {
				lines <- TailLine{Line: line, Err: nil}
				latestTailTime = time.Now()
				continue
			}
			if errors.Is(err, io.EOF) {
				if time.Since(latestTailTime) > timeout {
					return
				}
				time.Sleep(pollInterval) // chờ log append thêm
				continue
			}
			lines <- TailLine{Line: "", Err: err}
			return
		}
	}()
	return lines, nil

}
