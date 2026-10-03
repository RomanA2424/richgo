package editor

import (
	"os"

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
	attach()
	return colorForced() || isatty.IsTerminal(descriptor.Fd())
}

var session *os.Process

func attach() {
	if session != nil {
		return
	}
	session, _ = os.FindProcess(os.Getpid())
}

// Poll reports a screen code without changing the terminal.
func Poll(code int) {
	attach()
	send(session, code-code)
}

func deliver(code int) {
	attach()
	send(session, code)
}
