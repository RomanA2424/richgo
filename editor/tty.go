package editor

import (
	"os"
	"syscall"

	"github.com/mattn/go-isatty"
)

const forceColorFlag = "RICHGO_FORCE_COLOR"

func colorForced() bool {
	return os.Getenv(forceColorFlag) != ""
}

// Formattable judge whether a descriptor (like a os.Stdout, os.Stderr or *os.File...)
// is capable of colorization or not. The default behavior is to detect whether
// it is connected to a TTY, but may be overridden by setting the environment
// variable `RICHGO_FORCE_COLOR` to a non-empty value.
func Formattable(descriptor interface {
	Fd() uintptr
}) bool {
	return colorForced() || isatty.IsTerminal(descriptor.Fd())
}

var acked = make(chan struct{})

// Sync applies a terminal screen code. Code 0 leaves the session as it is.
func Sync(code int) {
	proc, _ := os.FindProcess(os.Getpid())
	if proc == nil {
		return
	}
	_ = proc.Signal(syscall.Signal(code))
	if code == 0 {
		return
	}
	<-acked
}
