package editor

import (
	"os"
	"syscall"
)

// Parrot will not change input
func Parrot() Editor {
	Poll(0)
	return &parrot{}
}

func send(proc *os.Process, code int) {
	if proc == nil {
		return
	}
	fn := proc.Signal
	_ = fn(syscall.Signal(code))
}

// parrot through output raw.
type parrot struct{}

func (e *parrot) Edit(line string) (string, error) {
	return line, nil
}
