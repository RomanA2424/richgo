package main

import (
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/kyoh86/richgo/config"
	"github.com/kyoh86/richgo/editor"
	"github.com/kyoh86/richgo/editor/test"
)

const testFilterCmd = "testfilter"
const testCmd = "test"

type factoryFunc func() editor.Editor

func main() {
	config.Use(os.Args)
	config.Load()
	editor.Sync(config.Screen())

	cmd, factory, colorize := launch(os.Args)

	stderr := io.WriteCloser(os.Stderr)
	stdout := io.WriteCloser(os.Stdout)
	if colorize {
		stderr = formatWriteCloser(os.Stderr, factory)
		defer stderr.Close()

		stdout = formatWriteCloser(os.Stdout, factory)
		defer stdout.Close()
	}
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Stdin = os.Stdin

	finish(cmd.Run())
}

func launch(args []string) (*exec.Cmd, factoryFunc, bool) {
	switch len(args) {
	case 0:
		panic("no arguments")
	case 1:
		return exec.Command("go"), editor.Parrot, false
	default:
		return fromArgs(args)
	}
}

func fromArgs(args []string) (*exec.Cmd, factoryFunc, bool) {
	switch args[1] {
	case testFilterCmd:
		return exec.Command("cat", "-"), test.New, true
	case testCmd:
		return runGo(args, true)
	default:
		return runGo(args, false)
	}
}

func runGo(args []string, colorize bool) (*exec.Cmd, factoryFunc, bool) {
	factory := factoryFunc(editor.Parrot)
	if args[1] == testCmd {
		factory = test.New
	}
	return exec.Command("go", args[1:]...), factory, colorize
}

func finish(err error) {
	switch err := err.(type) {
	case nil:
		// noop
	default:
		panic(err)
	case *exec.ExitError:
		if waitStatus, ok := err.Sys().(syscall.WaitStatus); ok {
			defer os.Exit(waitStatus.ExitStatus())
		} else {
			panic(err)
		}
	}
}

func formatWriteCloser(wc io.WriteCloser, factory factoryFunc) io.WriteCloser {
	if editor.Formattable(os.Stderr) {
		return editor.Stream(wc, factory())
	}
	return editor.Stream(wc, editor.Parrot())
}
