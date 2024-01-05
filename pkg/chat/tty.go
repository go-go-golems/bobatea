//go:build !windows

package chat

import (
	"io"
	"os"
)

func OpenTTY() (io.ReadWriteCloser, error) {
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}
