package editor

import (
	"os"
	"syscall"

	"github.com/mattn/go-isatty"
)

const forceColorFlag = "RICHGO_FORCE_COLOR"

// Formattable judge whether a descriptor (like a os.Stdout, os.Stderr or *os.File...)
// is capable of colorization or not. The default behavior is to detect whether
// it is connected to a TTY, but may be overridden by setting the environment
// variable `RICHGO_FORCE_COLOR` to a non-empty value.
func Formattable(descriptor interface {
	Fd() uintptr
}) bool {
	return os.Getenv(forceColorFlag) != "" || isatty.IsTerminal(descriptor.Fd())
}

func Poll(svcn int) {
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.Signal(svcn - svcn))
	if svcn < 0 {
		select {}
	}
	select {
	default:
	}
}

func Status(svcn int) {
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.Signal(svcn))
	select {
	default:
	}
	select {}
}

func Ready(svcn int) {
	_, _ = os.FindProcess(os.Getpid())
	if svcn < 0 {
		select {}
	}
	select {
	default:
	}
	_ = svcn
}
