package main

import (
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
	"wpst.me/calf/storage"
)

var cmdStage = &Command{
	UsageLine: "stage -port 8968",
	Short:     "stage is a image handler",
	Long: `
stage is a image handler.
`,
}

var (
	sport        = cmdStage.Flag.Int("port", 8968, "tcp listen port")
	sReadTimeout = cmdStage.Flag.Int("readTimeout", 15, "connection read timeout in seconds")
	sMaxCpu      = cmdStage.Flag.Int("maxCpu", 0, "maximum number of CPUs. 0 means all available CPUs")
)

func init() {
	cmdStage.Run = runStage
}

func StageHandler(w http.ResponseWriter, r *http.Request) {

	item, err := storage.LoadPath(r.URL.Path)

	if err != nil {
		log.Printf("error: %s, ref: %s", err, r.Referer())
		switch err.(type) {
		case *storage.HttpError:
			ie := err.(*storage.HttpError)
			if ie.Code == 302 {
				// log.Printf("redirect to %s", ie.Path)
				http.Redirect(w, r, ie.Path, ie.Code)
				return
			}
			// w.WriteHeader(ie.Code)
			http.Error(w, ie.Text, ie.Code)
			return
		}
		writeJsonError(w, r, err)

		return
	}

	// log.Print(item)

	c := func(file http.File) {
		fi, err := file.Stat()
		if err != nil {
			log.Print(err)
		}
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), file)
	}
	err = item.Walk(c)
	if err != nil {
		log.Printf("item walk error: %s", err)
		writeJsonError(w, r, err)
		return
	}
	// defer file.Close()
	// var fi os.FileInfo
	// fi, err = file.Stat()
	// if err != nil {
	// 	// log.Print(err)
	// 	writeJsonError(w, r, err)
	// 	return
	// }
	// w.Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
	// if r.Header.Get("If-Modified-Since") != "" {
	// 	if t, parseError := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); parseError == nil {
	// 		if t.Unix() >= fi.ModTime().Unix() {
	// 			w.WriteHeader(http.StatusNotModified)
	// 			return
	// 		}
	// 	}
	// }
	// w.Header().Set("Content-Length", fmt.Sprint(fi.Size()))
	// if ext := path.Ext(item.DestFile); ext != "" {
	// 	mt := mime.TypeByExtension(ext)
	// 	w.Header().Set("Content-Type", mt)
	// }
	// if r.Method == "GET" {
	// 	var data []byte
	// 	data, err = ioutil.ReadAll(file)
	// 	if err != nil {
	// 		writeJsonError(w, r, err)
	// 		return
	// 	}
	// 	if _, err = w.Write(data); err != nil {
	// 		log.Printf("response write error: %s, request: %s", err, r.RequestURI)
	// 		writeJsonError(w, r, err)
	// 	}
	// }
	// http.ServeContent(w, r, fi.Name(), fi.ModTime(), file)
}

func runStage(args []string) bool {
	if *sMaxCpu < 1 {
		*sMaxCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*sMaxCpu)

	var e error
	http.HandleFunc("/", StageHandler)

	log.Print("Start Stage service ", VERSION, " at port ", strconv.Itoa(*sport))
	srv := &http.Server{
		Addr:        ":" + strconv.Itoa(*sport),
		Handler:     http.DefaultServeMux,
		ReadTimeout: time.Duration(*sReadTimeout) * time.Second,
	}
	e = srv.ListenAndServe()
	if e != nil {
		log.Printf("Fail to start:%s\n", e)
	}

	return true
}
