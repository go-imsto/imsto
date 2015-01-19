package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
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
	m := newApiMeta(true)
	// m["roofs"] = config.Sections()
	writeJsonQuiet(w, r, newApiRes(m, config.Sections()))
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

	filter := storage.MetaFilter{Tags: r.FormValue("tags")}

	mw := storage.NewMetaWrapper(roof)
	t, err := mw.Count(filter)
	if err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	a, err := mw.Browse(int(limit), int(offset), sort, filter)
	if err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	m := newApiMeta(true)
	m["rows"] = limit
	m["page"] = page

	m["total"] = t

	// thumb_path := config.GetValue(roof, "thumb_path")
	// m["thumb_path"] = strings.TrimSuffix(thumb_path, "/") + "/"
	m["url_prefix"] = getUrl(r.URL.Scheme, roof, "") + "/"
	m["version"] = VERSION
	writeJsonQuiet(w, r, newApiRes(m, a))
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	roof := r.FormValue("roof")

	filter := storage.MetaFilter{Tags: r.FormValue("tags")}

	mw := storage.NewMetaWrapper(roof)
	t, err := mw.Count(filter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	m := newApiMeta(true)
	m["total"] = t
	m["version"] = VERSION
	writeJsonQuiet(w, r, newApiRes(m, nil))
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		err = fmt.Errorf("invalid path: %s", r.URL.Path)
		log.Print(err)
		writeJsonError(w, r, err)
		return
	}

	if err = r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("form parse error:", err)
		return
	}
	roof := parts[1]
	r.Form.Set("roof", roof)
	var id string
	if len(parts) > 2 {
		id = parts[2]
	}
	if id == "metas" {
		if len(parts) > 3 && parts[3] == "count" {
			countHandler(w, r)
			return
		}
		browseHandler(w, r)
		return
	}
	if id == "token" && r.Method == "POST" {
		tokenHandler(w, r)
		return
	}
	if id == "ticket" {
		ticketHandler(w, r)
		return
	}

	switch r.Method {
	case "GET", "HEAD":
		GetOrHeadHandler(w, r, roof, id)
	case "DELETE":
		secure(whiteList, DeleteHandler)(w, r)
	case "POST":
		secure(whiteList, PostHandler)(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJsonError(w, r, fmt.Errorf(http.StatusText(http.StatusMethodNotAllowed)))
	}
}

func GetOrHeadHandler(w http.ResponseWriter, r *http.Request, roof, ids string) {
	id, err := storage.NewEntryId(ids)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	mw := storage.NewMetaWrapper(roof)
	entry, err := mw.GetMeta(*id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	if r.Method == "HEAD" {
		return
	}
	url := getUrl(r.URL.Scheme, roof, "orig/"+entry.Path)
	log.Printf("Get entry: ", entry.Id)
	meta := newApiMeta(true)
	obj := struct {
		*storage.Entry
		OrigUrl string `json:"orig_url,omitempty"`
	}{
		Entry:   entry,
		OrigUrl: url,
	}
	writeJsonQuiet(w, r, newApiRes(meta, obj))
}

func getUrl(scheme, roof, size string) string {
	thumbPath := config.GetValue(roof, "thumb_path")
	spath := path.Join("/", thumbPath, size)
	stageHost := config.GetValue(roof, "stage_host")
	if stageHost == "" {
		return spath
	}
	if scheme == "" {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s%s", scheme, stageHost, spath)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := storage.StoredRequest(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}
	// log.Print(entries[0].Path)
	meta := newApiMeta(true)
	var roof = r.FormValue("roof")

	meta["stage_host"] = config.GetValue(roof, "stage_host")
	meta["url_prefix"] = getUrl(r.URL.Scheme, roof, "") + "/"
	meta["version"] = VERSION

	writeJsonQuiet(w, r, newApiRes(meta, entries))
}
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := storage.DeleteRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	meta := newApiMeta(true)
	writeJsonQuiet(w, r, newApiRes(meta, nil))
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := storage.TokenRequestNew(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJsonError(w, r, err)
		return
	}

	meta := newApiMeta(true)
	meta["token"] = token.String()
	writeJsonQuiet(w, r, newApiRes(meta, nil))
}

func ticketHandler(w http.ResponseWriter, r *http.Request) {
	meta := newApiMeta(false)
	if r.Method == "POST" {
		token, err := storage.TicketRequestNew(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("ERROR: %s", err)
			writeJsonError(w, r, err)
			return
		}
		meta["ok"] = true
		meta["token"] = token.String()
	} else if r.Method == "GET" {
		ticket, err := storage.TicketRequestLoad(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("ERROR: %s", err)
			writeJsonError(w, r, err)
			return
		}
		meta["ok"] = true
		meta["ticket"] = ticket
	}

	writeJsonQuiet(w, r, newApiRes(meta, nil))
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

	http.HandleFunc("/imsto/", storeHandler)
	http.HandleFunc("/imsto/meta", browseHandler)
	http.HandleFunc("/imsto/roofs", roofsHandler)
	http.HandleFunc("/imsto/token", tokenHandler)
	http.HandleFunc("/imsto/ticket", ticketHandler)

	// log.Print("Start Tiring service ", VERSION, " at port ", strconv.Itoa(*mport))
	str := fmt.Sprintf("Start Tiring service %s at port %d", VERSION, *mport)
	fmt.Println(str)
	log.Print(str)
	srv := &http.Server{
		Addr:        ":" + strconv.Itoa(*mport),
		Handler:     http.DefaultServeMux,
		ReadTimeout: time.Duration(*mReadTimeout) * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Printf("Fail to start: %s\n", err)
		return false
	}

	return true
}
