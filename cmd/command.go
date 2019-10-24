// Package cmd The command line tool for running Imsto bootstrap.
package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	// "runtime"
	"strings"
	"sync"
	"text/template"

	"go.uber.org/zap"

	"github.com/go-imsto/imsto/config"
	zlog "github.com/go-imsto/imsto/log"
	"github.com/go-imsto/imsto/storage"
)

// Command Cribbed from the genius organization of the "go" command.
type Command struct {
	Run                    func(args []string) bool
	UsageLine, Short, Long string
	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

func (cmd *Command) Name() string {
	name := cmd.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (cmd *Command) Usage() {
	fmt.Fprintf(os.Stderr, "Usage: imsto %s\n", cmd.UsageLine)
	fmt.Fprintf(os.Stderr, "Default Usage:\n")
	cmd.Flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Description:\n")
	fmt.Fprintf(os.Stderr, "  %s\n", strings.TrimSpace(cmd.Long))
	os.Exit(2)
}

// main
var (
	exitStatus = 0
	exitMu     sync.Mutex
)

var commands = []*Command{
	cmdAuth,
	// cmdImport,
	// cmdExport,
	// cmdOptimize,
	cmdRPC,
	cmdTiring,
	cmdStage,
	cmdView,
	cmdTest,
}

func setExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

func logger() zlog.Logger {
	return zlog.Get()
}

func init() {
	flag.Parse()

	storage.InitMetaTables()
}

func Main() {
	// fmt.Fprintf(os.Stdout, header)
	flag.Usage = func() { usage(1) }
	args := flag.Args()

	if len(args) < 1 || args[0] == "help" {
		if len(args) == 1 {
			usage(0)
		}
		if len(args) > 1 {
			for _, cmd := range commands {
				if cmd.Name() == args[1] {
					tmpl(os.Stdout, helpTemplate, cmd)
					return
				}
			}
		}
		usage(2)
	}

	var logger *zap.Logger
	if config.InDevelop() {
		logger, _ = zap.NewDevelopment()
		logger.Debug("logger start")
	} else {
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	zlog.Set(sugar)

	for _, cmd := range commands {
		name := cmd.Name()
		if name == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()

			// }
			if !cmd.Run(args) {
				fmt.Fprintf(os.Stderr, "\n")
				cmd.Flag.Usage()
				fmt.Fprintf(os.Stderr, "Default Parameters:\n")
				cmd.Flag.PrintDefaults()
			}
			// exit()
			return
		}
	}

	errorf("unknown command %q\nRun 'imsto help' for usage.\n", args[0])
}

func errorf(format string, args ...interface{}) {
	// Ensure the user's command prompt starts on the next line.
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

// const header = `
// Welcome

// `

const usageTemplate = `usage: imsto command [arguments]

The commands are:
{{range .}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use "imsto help [command]" for more information.
`

var helpTemplate = `usage: imsto {{.UsageLine}}
{{.Long}}
`

func usage(exitCode int) {
	fmt.Fprintln(os.Stderr, "version ", config.Version)
	tmpl(os.Stderr, usageTemplate, commands)
	os.Exit(exitCode)
}

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

var atExitFuncs []func()

func atExit(f func()) {
	atExitFuncs = append(atExitFuncs, f)
}

func exit() {
	for _, f := range atExitFuncs {
		f()
	}
	os.Exit(exitStatus)
}
