// The command line tool for running Imsto bootstrap.
package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	// "runtime"
	"strings"
	"sync"
	"text/template"

	"go.uber.org/zap"

	"github.com/go-imsto/imsto/config"
	zlog "github.com/go-imsto/imsto/log"
)

// Cribbed from the genius organization of the "go" command.
type Command struct {
	Run                    func(args []string) bool
	UsageLine, Short, Long string
	// Flag is a set of flags specific to this command.
	Flag    flag.FlagSet
	IsDebug bool
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

var (
	VERSION = "0.0.6"
)

// main
var (
	IsDebug    bool
	exitStatus = 0
	exitMu     sync.Mutex
	cfgDir     string
)

var commands = []*Command{
	// cmdImport,
	// cmdExport,
	// cmdOptimize,
	cmdTiring,
	cmdStage,
	cmdView,
	cmdTest,
	cmdAuth,
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
	flag.StringVar(&cfgDir, "conf", "/etc/imsto", "app config dir")
	flag.Parse()
	if cfgDir != "" {
		config.SetRoot(cfgDir)
	}
	err := config.Load()
	if err != nil {
		log.Print("config load error: ", err)
	}
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

	for _, cmd := range commands {
		name := cmd.Name()
		if name == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			IsDebug = cmd.IsDebug

			var logger *zap.Logger
			if IsDebug {
				logger, _ = zap.NewDevelopment()
				logger.Debug("logger start")
			} else {
				logger, _ = zap.NewProduction()
			}
			defer logger.Sync() // flushes buffer, if any
			sugar := logger.Sugar()

			zlog.Set(sugar)

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

type apiRes map[string]interface{}
type apiMeta map[string]interface{}
type apiError struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"message,omitempty"`
	err  error  `json:"-"`
}

func newApiRes(meta apiMeta, data interface{}) apiRes {
	res := make(apiRes)
	res["meta"] = meta
	res["data"] = data
	return res
}

func newApiMeta(ok bool) apiMeta {
	meta := make(apiMeta)
	meta["ok"] = ok
	return meta
}

func newApiError(err error) apiError {
	ae := apiError{err: err}
	ae.Msg = err.Error()
	return ae
}

func writeJson(w http.ResponseWriter, r *http.Request, obj interface{}) (err error) {
	w.Header().Set("Content-Type", "application/json")
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
		logger().Warnw("error writing JSON %s: %s", obj, err.Error())
	}
}

func writeJsonError(w http.ResponseWriter, r *http.Request, err error) {
	if r.Method == "GET" || r.Method == "HEAD" {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Pragma", "no-cache")
	}

	res := newApiRes(newApiMeta(false), nil)
	res["error"] = newApiError(err)

	writeJsonQuiet(w, r, res)
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
		w.WriteHeader(http.StatusForbidden)
		writeJsonQuiet(w, r, map[string]interface{}{"error": "No write permisson from " + host})
	}
}
