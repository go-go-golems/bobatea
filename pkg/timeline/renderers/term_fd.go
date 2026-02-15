package renderers

import (
	"math"
	"os"

	"golang.org/x/term"
)

func stdoutFD() (int, bool) {
	fd := os.Stdout.Fd()
	if fd > uintptr(math.MaxInt) {
		return 0, false
	}
	return int(fd), true
}

func stdoutIsTerminal() bool {
	fd, ok := stdoutFD()
	if !ok {
		return false
	}
	return term.IsTerminal(fd)
}
