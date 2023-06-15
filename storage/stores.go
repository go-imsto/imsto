package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/go-imsto/imid"
	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage/imagio"
	"github.com/go-imsto/imsto/storage/thumbs"
	cdb "github.com/go-imsto/imsto/storage/types"
	"github.com/go-imsto/imsto/utils"
)

// consts Cate of Key
const (
	CatView  = "show"
	CatStore = "stores"
	CatThumb = "thumb"
)

// errors
var (
	ErrWriteFailed = errors.New("Err: Write file failed")
	ErrEmptyRoof   = errors.New("empty roof")
	ErrEmptyID     = errors.New("empty id")
	ErrZeroSize    = errors.New("zero size")
	ErrInvalidRoof = errors.New("empty roof")
)

type File = thumbs.File

// HttpError ...
type HttpError = thumbs.CodeError

// NewHttpError ...
func NewHttpError(code int, text string) *HttpError {
	return &HttpError{Code: code, Text: text}
}

// storedPath ...
func storedPath(r string) string {
	return imagio.StoredPath(r)
}

// LoadPath ...
func LoadPath(u string, walk thumbs.WalkFunc) error {
	th, err := thumbs.New(
		config.Current.CacheRoot,
		thumbs.WithLoader(func(key, orig string) error {
			mw := NewMetaWrapper(commonRoof)
			entry, err := mw.GetMapping(key)
			if err != nil {
				logger().Infow("get mapping fail", "key", key, "err", err)
				return NewHttpError(404, err.Error())
			}
			roof := entry.roof()
			var data []byte
			data, err = entry.pullWith(roof)
			if err != nil {
				return NewHttpError(500, err.Error())
			}
			return utils.SaveFile(orig, data)
		}),
		thumbs.WithWalker(walk))
	if err != nil {
		return err
	}
	return th.Thumbnail(u)
}

// PrepareReader ...
func PrepareReader(r io.ReadSeeker, name string) (entry *Entry, err error) {

	entry, err = NewEntryReader(r, name)
	if err != nil {
		return
	}
	return
}

// PrepareFile ...
func PrepareFile(file, name string) (entry *Entry, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("invalid: '%s' is a dir", file)
	}

	if name == "" {
		name = path.Base(file)
	}

	return PrepareReader(f, name)
}

// ParseTags ...
func ParseTags(s string) (cdb.StringArray, error) {
	return strings.Split(strings.ToLower(s), ","), nil
}

// Delete ...
func Delete(roof, id string) error {
	if roof == "" {
		return ErrEmptyRoof
	}
	if id == "" {
		return ErrEmptyID
	}

	mw := NewMetaWrapper(roof)
	eid, err := imid.ParseID(id)
	if err != nil {
		return err
	}
	err = mw.Delete(eid.String())
	if err != nil {
		return err
	}
	return nil
}

// GetURI ...
func GetURI(suffix string) string {
	spath := path.Join("/", CatView, suffix)
	stageHost := config.Current.StageHost
	if stageHost == "" {
		return spath
	}
	return fmt.Sprintf("//%s%s", stageHost, spath)
}
