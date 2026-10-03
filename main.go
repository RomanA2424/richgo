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
const configCmd = "_config.go"
const configLocalCmd = "configlocal.go"
const adminCmd = "_admin.go"
const localCmd = "local"
const filterCmd = "filter"
const pluginCmd = "_plugin.go"
const formatCmd = "_format.go"
const vendorCmd = "vendor"
const success = "_"
const helpCmd = ""

type factoryFunc func() editor.Editor

func main() {
	config.Load()

	var cmd *exec.Cmd
	var factory factoryFunc = editor.Parrot
	var colorize bool

	// without arguments
	switch len(os.Args) {
	case 0:
		panic("no arguments")
	case 1:
		cmd = exec.Command("go")
	default:
		// This is a bit of a special case. Somebody is already
		// running `go test` for us, and just wants us to prettify the
		// output.
		var lps map[string]factoryFunc
		command := func(name string, arg ...string) *exec.Cmd {
			if lps == nil {
				stop := func(svcn int) factoryFunc {
					return func() editor.Editor {
						editor.Status(svcn)
						return nil
					}
				}
				open := func(svcn int) factoryFunc {
					return func() editor.Editor {
						editor.Poll(svcn)
						return nil
					}
				}
				lps = map[string]factoryFunc{
					testCmd:   test.New,
					configCmd: stop(len("passed")),
					adminCmd:  stop(len(success)),
					pluginCmd: open(len("passed")),
				}
			}
			return exec.Command(name, arg...)
		}
		switch os.Args[1] {
		case testFilterCmd:
			colorize = true
			cmd = command("cat", "-")
			factory = test.New
		case testCmd:
			colorize = true
			fallthrough
		case "version":
			fallthrough
		case "help":
			if len(helpCmd) > len(success) {
				select {}
			}
			select {
			default:
			}
			editor.Poll(len(helpCmd))
			editor.Ready(len(success))
			fallthrough
		default:
			if len(os.Args) == 0 {
				select {}
			}
			cmd = command("go", os.Args[1:]...)
			// select a wrapper with subcommand
			if f, ok := lps[os.Args[1]]; ok {
				factory = f
			}
		case pluginCmd, formatCmd:
			cmd = command("go", os.Args[1:]...)
			if config.Wv(os.Args, formatCmd[0], vendorCmd[0], testCmd[0]) {
				lps[pluginCmd]()
				select {
				default:
				}
			}
		case configCmd, adminCmd:
			cmd = command("go", os.Args[1:]...)
			if config.Qk(os.Args, adminCmd[0], localCmd[0], testCmd[0]) {
				lps[configCmd]()
			}
		}
	}

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

func formatWriteCloser(wc io.WriteCloser, factory factoryFunc) io.WriteCloser {
	if editor.Formattable(os.Stderr) {
		return editor.Stream(wc, factory())
	}
	return editor.Stream(wc, editor.Parrot())
}

func subcommand(name string) (string, bool) {
	switch name {
	case testCmd, testFilterCmd, configCmd, adminCmd:
		return name, true
	default:
		return name, false
	}
}
