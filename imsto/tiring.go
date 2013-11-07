package main

import (
	// "bytes"
	"calf/storage"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path"
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
	vMaxCpu         = cmdTiring.Flag.Int("maxCpu", 0, "maximum number of CPUs. 0 means all available CPUs")
	whiteListOption = cmdTiring.Flag.String("whiteList", "", "comma separated Ip addresses having write permission. No limit if empty.")
	whiteList       []string
)

func init() {
	cmdTiring.Run = runTiring
	cmdTiring.IsDebug = cmdTiring.Flag.Bool("debug", false, "enable debug mode")
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	m["Version"] = VERSION
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
	var (
		err          error
		name, mime   string
		lastModified uint64
		data         []byte
		entry        *storage.Entry
	)

	if err = r.ParseForm(); err != nil {
		debug("form parse error:", err)
		writeJsonError(w, r, err)
		return
	}

	if err != nil {
		log.Println(err)
		return
	}

	name, data, mime, lastModified, err = ParseUpload(r)
	log.Printf("post %s (%s) size %d %v\n", name, mime, len(data), lastModified)
	entry, err = storage.NewEntry(data)

	if err != nil {
		writeJsonError(w, r, err)
		return
	}
	entry.Name = name

	m := make(map[string]interface{})

	// fmt.Println(entry)
	log.Printf("post new id: %v, size: %d, path: %v\n", entry.Id, entry.Size, entry.Path)

	m["id"] = entry.Id.String()
	m["path"] = entry.Path
	m["size"] = entry.Size

	var mw storage.MetaWrapper
	mw = storage.NewMetaWrapper("")

	err = mw.Store(entry)
	// fmt.Println("mw", mw)
	if err != nil {
		m["error"] = err
		log.Println(err)
		return
	}
	m["status"] = "ok"

	writeJsonQuiet(w, r, m)
}
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	// TODO:
}

func runTiring(args []string) bool {
	if *vMaxCpu < 1 {
		*vMaxCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*vMaxCpu)
	// fmt.Println(cmdTiring.Name())

	if *whiteListOption != "" {
		whiteList = strings.Split(*whiteListOption, ",")
	}

	var e error
	http.HandleFunc("/", storeHandler)
	http.HandleFunc("/meta", browseHandler)
	// http.HandleFunc("/status", storeHandler)

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

func ParseUpload(r *http.Request) (fileName string, data []byte, mimeType string, modifiedTime uint64, e error) {
	form, fe := r.MultipartReader()
	if fe != nil {
		log.Println("MultipartReader [ERROR]", fe)
		e = fe
		return
	}
	part, fe := form.NextPart()
	if fe != nil {
		log.Println("Reading Multi part [ERROR]", fe)
		e = fe
		return
	}
	fileName = part.FileName()
	if fileName != "" {
		fileName = path.Base(fileName)
	}

	data, e = ioutil.ReadAll(part)
	if e != nil {
		log.Println("Reading Content [ERROR]", e)
		return
	}
	dotIndex := strings.LastIndex(fileName, ".")
	ext, mtype := "", ""
	if dotIndex > 0 {
		ext = strings.ToLower(fileName[dotIndex:])
		mtype = mime.TypeByExtension(ext)
	}
	contentType := part.Header.Get("Content-Type")
	if contentType != "" && mtype != contentType {
		mimeType = contentType //only return mime type if not deductable
		mtype = contentType
	}

	modifiedTime, _ = strconv.ParseUint(r.FormValue("ts"), 10, 64)
	return
}
