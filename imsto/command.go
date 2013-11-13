// The command line tool for running Revel apps.
package main

import (
	"calf/config"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
)

// Cribbed from the genius organization of the "go" command.
type Command struct {
	Run                    func(args []string) bool
	UsageLine, Short, Long string
	// Flag is a set of flags specific to this command.
	Flag    flag.FlagSet
	IsDebug *bool
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
	fmt.Fprintf(os.Stderr, "Example: imsto %s\n", cmd.UsageLine)
	fmt.Fprintf(os.Stderr, "Default Usage:\n")
	cmd.Flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Description:\n")
	fmt.Fprintf(os.Stderr, "  %s\n", strings.TrimSpace(cmd.Long))
	os.Exit(2)
}

type LoggedError struct{ error }

const (
	VERSION = "0.0.2"
)

// main
var (
	IsDebug    *bool
	exitStatus = 0
	exitMu     sync.Mutex
	appDir     string
	logDir     string
)

var commands = []*Command{
	cmdImport,
	cmdOptimize,
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

func init() {
	flag.StringVar(&appDir, "root", "", "app root dir")
	flag.StringVar(&logDir, "logs", "", "app logs dir")
	flag.Parse()
	if appDir != "" {
		config.SetAppRoot(appDir)
	}
	err := config.Load()
	if err != nil {
		log.Print("config load error: ", err)
	}
	if logDir == "" {
		dir := config.AppRoot()
		log.Printf("approot %s", dir)
		if err != nil {
			log.Print("getwd error: ", err)
		}
		if dir != "" {
			dir = path.Join(dir, "logs")
			// log.Printf("log dir: %s", dir)
			if fi, err := os.Stat(dir); err == nil {
				if fi.IsDir() {
					logDir = dir
				}
			}
		}
	}
}

func main() {
	fmt.Fprintf(os.Stdout, header)
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

	for _, cmd := range commands {
		name := cmd.Name()
		if name == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			IsDebug = cmd.IsDebug
			// if IsDebug != nil && *IsDebug {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("log dir: %s", logDir)
			if logDir != "" {
				logfile := path.Join(logDir, name+".log")
				log.Printf("logfile: %s", logfile)
				fd, err := os.OpenFile(logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664)
				if err != nil {
					log.Printf("logfile %s create failed", logfile)
					return
				}
				log.SetOutput(fd)
			}

			// }
			if cmd.Run(args) {
				fmt.Fprintf(os.Stderr, "\n")
			}
			exit()
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
	// panic(LoggedError{}) // Panic instead of os.Exit so that deferred will run.
}

const header = `
Welcome

`

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

func writeJson(w http.ResponseWriter, r *http.Request, obj interface{}) (err error) {
	w.Header().Set("Content-Type", "application/javascript")
	var bytes []byte
	if r.FormValue("pretty") != "" {
		bytes, err = json.MarshalIndent(obj, "", "  ")
	} else {
		bytes, err = json.Marshal(obj)
	}
	if err != nil {
		return
	}
	callback := r.FormValue("callback")
	if callback == "" {
		_, err = w.Write(bytes)
	} else {
		if _, err = w.Write([]uint8(callback)); err != nil {
			return
		}
		if _, err = w.Write([]uint8("(")); err != nil {
			return
		}
		fmt.Fprint(w, string(bytes))
		if _, err = w.Write([]uint8(")")); err != nil {
			return
		}
	}
	return
}

// wrapper for writeJson - just logs errors
func writeJsonQuiet(w http.ResponseWriter, r *http.Request, obj interface{}) {
	if err := writeJson(w, r, obj); err != nil {
		log.Printf("error writing JSON %s: %s", obj, err.Error())
	}
}
func writeJsonError(w http.ResponseWriter, r *http.Request, err error) {
	m := make(map[string]interface{})
	m["status"] = "fail"
	m["error"] = err.Error()
	writeJsonQuiet(w, r, m)
}

func debug(params ...interface{}) {
	if *IsDebug {
		log.Print(params)
	}
}
func secure(whiteList []string, f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(whiteList) == 0 {
			f(w, r)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			for _, ip := range whiteList {
				if ip == host {
					f(w, r)
					return
				}
			}
		}
		writeJsonQuiet(w, r, map[string]interface{}{"error": "No write permisson from " + host})
	}
}
