package main

import (
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
	"wpst.me/calf/config"
	"wpst.me/calf/storage"
)

var cmdTiring = &Command{
	UsageLine: "tiring -port 8964 -whiteList=\"127.0.0.1\"",
	Short:     "serve tiring http service",
	Long: `
serve tiring http service
`,
}

var (
	mport           = cmdTiring.Flag.Int("port", 8964, "tcp listen port")
	mReadTimeout    = cmdTiring.Flag.Int("readTimeout", 3, "connection read timeout in seconds")
	mMaxCpu         = cmdTiring.Flag.Int("maxCpu", 0, "maximum number of CPUs. 0 means all available CPUs")
	whiteListOption = cmdTiring.Flag.String("whiteList", "", "comma separated Ip addresses having write permission. No limit if empty.")
	whiteList       []string
)

func init() {
	cmdTiring.Run = runTiring
	cmdTiring.IsDebug = cmdTiring.Flag.Bool("debug", false, "enable debug mode")
}

func roofsHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	m["roofs"] = config.Sections()
	writeJsonQuiet(w, r, m)
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	roof := r.FormValue("roof")
	// log.Printf("browse roof: %s", roof)
	var (
		limit uint64
		page  uint64
	)

	if str := r.FormValue("rows"); str != "" {
		limit, _ = strconv.ParseUint(str, 10, 32)
		if limit < 1 {
			limit = 1
		}
	} else {
		limit = 20
	}

	if str := r.FormValue("page"); str != "" {
		page, _ = strconv.ParseUint(str, 10, 32)
	}
	if page < 1 {
		page = 1
	}

	offset := limit * (page - 1)

	sort := make(map[string]int)
	sort_name := r.FormValue("sort_name")
	sort_order := r.FormValue("sort_order")
	if sort_name != "" {
		var o int
		if strings.ToUpper(sort_order) == "ASC" {
			o = storage.ASCENDING
		} else {
			o = storage.DESCENDING
		}
		sort[sort_name] = o
	}

	mw := storage.NewMetaWrapper(roof)
	t, err := mw.Count()
	if err != nil {
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}
	a, err := mw.Browse(int(limit), int(offset), sort)
	if err != nil {
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}
	m := make(map[string]interface{})
	m["rows"] = limit
	m["page"] = page

	m["data"] = a
	m["total"] = t

	thumb_path := config.GetValue(roof, "thumb_path")
	m["thumb_path"] = strings.TrimSuffix(thumb_path, "/") + "/"
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
	// log.Print(entries[0].Path)
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
	http.HandleFunc("/imsto/roofs", roofsHandler)
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
