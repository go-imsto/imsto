package storage

import (
	"calf/db"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

func LoadRequest(r *http.Request) error {
	// TODO:
	return nil
}

func StoredFile(filename string, section string) (entry *Entry, err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			log.Println(err)
			return
		}
	}

	var data []byte

	data, err = ioutil.ReadFile(filename)

	if err != nil {
		log.Println(err)
		return
	}

	entry, err = newEntry(data, path.Base(filename))

	if err != nil {
		return
	}
	entry.Modified = uint64(fi.ModTime().Unix())
	// err = entry.trek(section)

	err = store(entry, section)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func StoredRequest(r *http.Request) (entry *Entry, err error) {
	var (
		name, mime   string
		data         []byte
		lastModified uint64
	)

	if err = r.ParseForm(); err != nil {
		log.Print("form parse error:", err)
		return
	}

	if err != nil {
		log.Println(err)
		return
	}

	name, data, mime, lastModified, err = ParseUpload(r)
	log.Printf("post %s (%s) size %d %v\n", name, mime, len(data), lastModified)
	entry, err = newEntry(data, name)

	if err != nil {
		return
	}
	entry.Modified = lastModified
	section := r.FormValue("section")
	// err = entry.Trek(section)

	err = store(entry, section)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func store(e *Entry, section string) (err error) {
	err = e.Trek(section)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("new id: %v, size: %d, path: %v\n", e.Id, e.Size, e.Path)

	data := e.Blob()
	log.Printf("blob length: %d", len(data))

	mw := NewMetaWrapper(section)

	var em Wagoner
	em, err = FarmEngine(section)
	if err != nil {
		log.Println(err)
		return
	}

	var sev db.Hstore
	sev, err = em.Put(e.Path, data, e.Mime)
	if err != nil {
		log.Println(err)
		return
	}

	e.sev = sev

	err = mw.Store(e)
	// fmt.Println("mw", mw)
	if err != nil {
		log.Println(err)
		return
	}
	return
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
