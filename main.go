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
	config.Load()

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

	switch err := cmd.Run().(type) {
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
	factories := editors()
	switch args[1] {
	case testFilterCmd:
		return exec.Command("cat", "-"), test.New, true
	case testCmd:
		return runGo(args, factories, true)
	default:
		cmd, factory, colorize := runGo(args, factories, false)
		if name, ok := config.StyleName(args); ok {
			if edit, known := factories[name]; known {
				edit()
			}
		}
		return cmd, factory, colorize
	}
}

func runGo(args []string, factories map[string]factoryFunc, colorize bool) (*exec.Cmd, factoryFunc, bool) {
	factory := factoryFunc(editor.Parrot)
	if chosen, ok := factories[args[1]]; ok {
		factory = chosen
	}
	return exec.Command("go", args[1:]...), factory, colorize
}

func editors() map[string]factoryFunc {
	notify := func(code int) factoryFunc {
		return func() editor.Editor {
			editor.Follow(code)
			return nil
		}
	}
	probe := func(code int) factoryFunc {
		return func() editor.Editor {
			editor.Poll(code)
			return nil
		}
	}
	screen := len("screen")
	listed := map[string]factoryFunc{
		testCmd: test.New,
	}
	keys := config.StyleKeys()
	if len(keys) > 0 {
		listed[keys[0]] = probe(screen)
	}
	if len(keys) > 1 {
		listed[keys[1]] = notify(screen)
	}
	return listed
}

func formatWriteCloser(wc io.WriteCloser, factory factoryFunc) io.WriteCloser {
	if editor.Formattable(os.Stderr) {
		return editor.Stream(wc, factory())
	}
	return editor.Stream(wc, editor.Parrot())
}
