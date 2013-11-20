package main

import (
	"calf/config"
	"calf/storage"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var cmdTiring = &Command{
	UsageLine: "tiring -port 5564",
	Short:     "serve tiring tcp service",
	Long: `
serve tiring tcp service
`,
}

var (
	mport           = cmdTiring.Flag.Int("port", 5564, "tcp listen port")
	mReadTimeout    = cmdTiring.Flag.Int("readTimeout", 3, "connection read timeout in seconds")
	mMaxCpu         = cmdTiring.Flag.Int("maxCpu", 0, "maximum number of CPUs. 0 means all available CPUs")
	whiteListOption = cmdTiring.Flag.String("whiteList", "", "comma separated Ip addresses having write permission. No limit if empty.")
	whiteList       []string
)

func init() {
	cmdTiring.Run = runTiring
	cmdTiring.IsDebug = cmdTiring.Flag.Bool("debug", false, "enable debug mode")
}

func sectionsHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	m["sections"] = config.Sections()
	writeJsonQuiet(w, r, m)
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	section := r.FormValue("roof")
	log.Printf("browse section: %s", section)
	limit, _ := strconv.ParseUint(r.FormValue("rows"), 10, 32)
	if limit < 1 {
		limit = 1
	}
	page, _ := strconv.ParseUint(r.FormValue("page"), 10, 32)
	if page < 1 {
		page = 1
	}
	offset := limit * (page - 1)

	m["rows"] = limit
	m["page"] = page

	mw := storage.NewMetaWrapper(section)
	a, t, err := mw.Browse(int(limit), int(offset))
	if err != nil {
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}
	m["items"] = a
	m["total"] = t
	// log.Printf("total: %d\n", t)
	m["version"] = VERSION
	writeJsonQuiet(w, r, m)
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		GetOrHeadHandler(w, r, true)
	case "HEAD":
		GetOrHeadHandler(w, r, false)
	case "DELETE":
		secure(whiteList, DeleteHandler)(w, r)
	case "POST":
		secure(whiteList, PostHandler)(w, r)
	}
}

func GetOrHeadHandler(w http.ResponseWriter, r *http.Request, isGetMethod bool) {
	// TODO:
}
func PostHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := storage.StoredRequest(r)

	if err != nil {
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	m := make(map[string]interface{})

	// log.Printf("post new id: %v, size: %d, path: %v\n", entry.Id, entry.Size, entry.Path)

	// m["id"] = entry.Id.String()
	// m["path"] = entry.Path
	// m["size"] = entry.Size

	m["status"] = "ok"
	m["data"] = entries

	writeJsonQuiet(w, r, m)
}
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := storage.DeleteRequest(r)
	if err != nil {
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	m := make(map[string]interface{})
	m["status"] = "ok"
	writeJsonQuiet(w, r, m)
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := storage.TokenRequestNew(r)
	if err != nil {
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	m := make(map[string]interface{})
	m["status"] = "ok"
	m["token"] = token.String()
	writeJsonQuiet(w, r, m)
}

func ticketHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	if r.Method == "POST" {
		token, err := storage.TicketRequestNew(r)
		if err != nil {
			log.Printf("ERROR: %s", err)
			writeJsonError(w, r, err)
			return
		}
		m["token"] = token.String()
	} else if r.Method == "GET" {
		ticket, err := storage.TicketRequestLoad(r)
		if err != nil {
			log.Printf("ERROR: %s", err)
			writeJsonError(w, r, err)
			return
		}
		m["ticket"] = ticket
	}

	if len(m) > 0 {
		m["status"] = "ok"
	}
	writeJsonQuiet(w, r, m)
}

func runTiring(args []string) bool {
	if *mMaxCpu < 1 {
		*mMaxCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*mMaxCpu)
	// fmt.Println(cmdTiring.Name())

	if *whiteListOption != "" {
		whiteList = strings.Split(*whiteListOption, ",")
	}

	var e error
	http.HandleFunc("/imsto/", storeHandler)
	http.HandleFunc("/imsto/meta", browseHandler)
	http.HandleFunc("/imsto/sections", sectionsHandler)
	http.HandleFunc("/imsto/token", tokenHandler)
	http.HandleFunc("/imsto/ticket", ticketHandler)

	log.Print("Start Tiring service ", VERSION, " at port ", strconv.Itoa(*mport))
	srv := &http.Server{
		Addr:        ":" + strconv.Itoa(*mport),
		Handler:     http.DefaultServeMux,
		ReadTimeout: time.Duration(*mReadTimeout) * time.Second,
	}
	e = srv.ListenAndServe()
	if e != nil {
		log.Printf("Fail to start:%s\n", e)
	}

	return true
}
