package web

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/bmizerany/pat"

	"github.com/go-imsto/imagid"
	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage"
)

// Handler ...
func Handler() http.Handler {
	mux := pat.New()
	mux.Get("/imsto/roofs", http.HandlerFunc(roofsHandler))

	mux.Post("/imsto/ticket", storage.CheckAPIKey(http.HandlerFunc(ticketHandlerPost)))
	mux.Get("/imsto/ticket", storage.CheckAPIKey(http.HandlerFunc(ticketHandlerGet)))

	mux.Post("/imsto/token", storage.CheckAPIKey(http.HandlerFunc(tokenHandler)))

	mux.Post("/imsto/:roof", storage.CheckAPIKey(secure(storedHandler)))
	mux.Del("/imsto/:roof/:id", storage.CheckAPIKey(secure(deleteHandler)))
	mux.Get("/imsto/:roof/id", http.HandlerFunc(GetOrHeadHandler))
	mux.Get("/imsto/:roof/metas/count", http.HandlerFunc(countHandler))
	mux.Get("/imsto/:roof/metas", http.HandlerFunc(browseHandler))
	// mux.Post("/imsto/:roof/token", http.HandlerFunc(tokenHandler))
	// mux.Post("/imsto/:roof/ticket", http.HandlerFunc(ticketHandler))

	return mux
}

// StageHandler ...
func StageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Server", "IMSTO STAGE")

	item, err := storage.LoadPath(r.URL.Path)

	if err != nil {
		logger().Warnw("loadPath fail", "uri", r.URL.Path, "ref", r.Referer(), "err", err)
		if he, ok := err.(*storage.HttpError); ok {
			if he.Code == 302 {
				logger().Infow("redirect", "path", he.Path)
				http.Redirect(w, r, he.Path, he.Code)
				return
			}
			w.WriteHeader(he.Code)
			writeJSONError(w, r, err)
			return
		}
		w.WriteHeader(400)
		writeJSONError(w, r, err)

		return
	}

	// log.Print(item)

	c := func(file storage.File) {
		http.ServeContent(w, r, file.Name(), file.Modified(), file)
	}
	err = item.Walk(c)
	if err != nil {
		logger().Warnw("item walk fail", "item", item, "err", err)
		w.WriteHeader(500)
		writeJSONError(w, r, err)
		return
	}
}

func roofsHandler(w http.ResponseWriter, r *http.Request) {
	m := newApiMeta(true)
	writeJSONQuiet(w, r, newApiRes(m, config.Current.Roofs))
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	roof := r.URL.Query().Get(":roof")
	logger().Debugw("browse", "roof", roof, "query", r.URL.Query())
	var (
		limit  uint64
		offset uint64
	)

	if str := r.FormValue("rows"); str != "" {
		limit, _ = strconv.ParseUint(str, 10, 32)
		if limit < 1 {
			limit = 1
		}
	} else {
		limit = 20
	}

	if str := r.FormValue("skip"); str != "" {
		offset, _ = strconv.ParseUint(str, 10, 32)
	} else {
		var page uint64
		if str := r.FormValue("page"); str != "" {
			page, _ = strconv.ParseUint(str, 10, 32)
		}
		if page < 1 {
			page = 1
		}

		offset = limit * (page - 1)
	}

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
		logger().Infow("count fail", "uri", r.RequestURI, "roof", roof, "filter", filter, "err", err)
		writeJSONError(w, r, err)
		return
	}

	a, err := mw.Browse(int(limit), int(offset), sort, filter)
	if err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJSONError(w, r, err)
		return
	}

	m := newApiMeta(true)
	m["rows"] = limit
	m["page"] = (offset + 1)
	m["skip"] = offset
	m["page_count"] = uint(math.Ceil(float64(t) / float64(limit)))

	m["total"] = t

	m["stageHost"] = config.Current.StageHost
	m["urlPrefix"] = getURL(r, "") + "/"
	m["version"] = config.Version
	writeJSONQuiet(w, r, newApiRes(m, a))
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	roof := r.FormValue("roof")

	filter := storage.MetaFilter{Tags: r.FormValue("tags")}

	mw := storage.NewMetaWrapper(roof)
	t, err := mw.Count(filter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ERROR: %s", err)
		writeJSONError(w, r, err)
		return
	}

	m := newApiMeta(true)
	m["total"] = t
	m["version"] = config.Version
	writeJSONQuiet(w, r, newApiRes(m, nil))
}

// GetOrHeadHandler ...
func GetOrHeadHandler(w http.ResponseWriter, r *http.Request) {
	id, err := imagid.ParseID(r.URL.Query().Get(":id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("ERROR: %s", err)
		writeJSONError(w, r, err)
		return
	}

	roof := r.URL.Query().Get(":roof")
	mw := storage.NewMetaWrapper(roof)
	entry, err := mw.GetMeta(id.String())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("ERROR: %s", err)
		writeJSONError(w, r, err)
		return
	}

	if r.Method == "HEAD" {
		return
	}
	url := getURL(r, "orig/"+entry.Path)
	log.Printf("Get entry: %v", entry.Id)
	meta := newApiMeta(true)
	obj := struct {
		*storage.Entry
		OrigURL string `json:"orig_url,omitempty"`
	}{
		Entry:   entry,
		OrigURL: url,
	}
	writeJSONQuiet(w, r, newApiRes(meta, obj))
}

func getURL(r *http.Request, size string) string {
	return storage.GetURI(size)
}

func storedHandler(w http.ResponseWriter, r *http.Request) {
	var us uploadSchema
	err := Bind(r, &us)
	if err != nil {
		w.WriteHeader(400)
		writeJSONError(w, r, err)
		return
	}
	if us.Roof == "" {
		us.Roof = r.URL.Query().Get(":roof")
	}
	if err = r.ParseMultipartForm(storage.DefaultMaxMemory); err != nil {
		log.Print("multipart form parse error:", err)
		w.WriteHeader(400)
		writeJSONError(w, r, err)
		return
	}
	app, appOK := storage.AppFromContext(r.Context())
	if !appOK {
		w.WriteHeader(400)
		writeJson(w, r, "app error")
		return
	}
	_, err = app.VerifyToken(us.Token)
	if err != nil {
		writeJSONError(w, r, err)
		return
	}

	tags, _ := storage.ParseTags(us.Tags)
	entries := make(map[string][]*storage.Entry)
	for k, fhs := range r.MultipartForm.File {
		entries[k] = make([]*storage.Entry, len(fhs))
		for i, fh := range fhs {
			entries[k][i] = new(storage.Entry)
			log.Printf("%d name: %s, ctype: %s", i, fh.Filename, fh.Header.Get("Content-Type"))
			mime := fh.Header.Get("Content-Type")
			file, fe := fh.Open()
			if fe != nil {
				entries[k][i].Err = fe.Error()
			}

			logger().Infow("post upload", "name", fh.Filename, "mime", mime, "size", fh.Size)

			entry, ee := storage.PrepareReader(file, fh.Filename)
			if ee != nil {
				logger().Infow("prepare upload fail", "name", fh.Filename)
				entries[k][i].Err = ee.Error()
				continue
			}
			entry.AppId = app.Id
			entry.Author = storage.Author(us.User)
			// entry.Modified = lastModified
			entry.Tags = tags
			ee = <-entry.Store(us.Roof)
			if ee != nil {
				logger().Infow("stored fail", "i", i, "roof", us.Roof, "id", entry.Id, "err", ee)
				entries[k][i].Err = ee.Error()
				continue
			}
			logger().Infow("stored", "i", i, "roof", us.Roof, "id", entry.Id, "path", entry.Path)
			entries[k][i] = entry
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJSONError(w, r, err)
		return
	}
	// log.Print(entries[0].Path)
	meta := newApiMeta(true)

	meta["stageHost"] = config.Current.StageHost
	meta["urlPrefix"] = getURL(r, "") + "/"
	meta["version"] = config.Version

	writeJSONQuiet(w, r, newApiRes(meta, entries))
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	err := storage.Delete(r.URL.Query().Get(":roof"), r.URL.Query().Get(":id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJSONError(w, r, err)
		return
	}

	meta := newApiMeta(true)
	writeJSONQuiet(w, r, newApiRes(meta, nil))
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	var param tokenSchema
	if err := Bind(r, &param); err != nil {
		writeJSONError(w, r, err)
		return
	}
	app, appOK := storage.AppFromContext(r.Context())
	if !appOK {
		w.WriteHeader(400)
		writeJson(w, r, "app error")
		return
	}

	token := app.RequestNewToken(param.User)
	meta := newApiMeta(token != "")
	meta["token"] = token
	writeJSONQuiet(w, r, newApiRes(meta, nil))
}

func ticketHandlerPost(w http.ResponseWriter, r *http.Request) {
	var param ticketSchema
	if err := Bind(r, &param); err != nil {
		writeJSONError(w, r, err)
		return
	}
	app, appOK := storage.AppFromContext(r.Context())
	if !appOK {
		w.WriteHeader(400)
		writeJson(w, r, "app error")
		return
	}

	token, err := app.TicketRequestNew(param.Roof, param.Token, param.User, param.Prompt)
	if err != nil {
		writeJSONError(w, r, err)
		return
	}
	str := token.String()
	meta := newApiMeta(str != "")
	meta["token"] = str
	writeJSONQuiet(w, r, newApiRes(meta, nil))
}

func ticketHandlerGet(w http.ResponseWriter, r *http.Request) {
	var param ticketSchema
	if err := Bind(r, &param); err != nil {
		writeJSONError(w, r, err)
		return
	}
	app, appOK := storage.AppFromContext(r.Context())
	if !appOK {
		w.WriteHeader(400)
		writeJson(w, r, "app error")
		return
	}
	ticket, err := app.TicketRequestLoad(param.Token, param.Roof, param.User)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR: %s", err)
		writeJSONError(w, r, err)
		return
	}
	meta := newApiMeta(false)
	meta["ok"] = true
	meta["ticket"] = ticket

	writeJSONQuiet(w, r, newApiRes(meta, nil))
}
