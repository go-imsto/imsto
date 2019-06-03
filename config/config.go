package config

import (
	"errors"
	"github.com/vaughan0/go-ini"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

const defaultConfigIni = `[common]
;meta_table_suffix = demo
engine = s3
bucket_name = imsto-demo
max_quality = 88
max_file_size = 262114
max_width = 1600
max_height = 1600
min_width = 50
min_height = 50
thumb_path = /thumb
thumb_root = /opt/imsto/cache/thumb/
temp_root = /tmp/
support_size = 120,160,400
ticket_table = upload_ticket
watermark = watermark.png
watermark_opacity = 20
;copyright_label = imsto.net
copyright =
log_dir = /var/log/imsto
stage_host =
`

// var once sync.Once

var (
	cfgDir        string
	logDir        string
	defaultConfig ini.File
	loadedConfig  ini.File
)

func Root() string {
	return cfgDir
}

func SetRoot(dir string) {

	if _, err := os.Stat(dir); err != nil {
		log.Println(err)
		return
	}

	cfgDir = dir
}

func init() {
	defaultConfig, _ = ini.Load(strings.NewReader(defaultConfigIni))
}

func GetValue(section, name string) string {
	var (
		s  string
		ok bool
	)

	if s, ok = loadedConfig.Get(section, name); !ok {
		if s, ok = loadedConfig.Get("common", name); !ok {
			if s, ok = defaultConfig.Get("common", name); !ok {
				log.Printf("'%v' variable missing from '%v' section", name, section)
				return ""
			}
		}
	}

	return strings.Trim(s, "\"")
}

func GetInt(section, name string) int {
	s := GetValue(section, name)
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return int(i)
	}
	return 0
}

func GetSection(sname string) (section ini.Section) {
	if section = loadedConfig.Section(sname); len(section) > 0 {
		return section
	}

	return defaultConfig.Section("")
}

func HasSection(sname string) bool {
	if _, ok := loadedConfig[sname]; ok {
		return true
	}
	return false
}

// return administrable sections
func Sections() map[string]string {
	a := make(map[string]string)
	for k, v := range loadedConfig {
		if k != "common" {
			// a = append(a, k)
			if admin, oka := v["administrable"]; oka && admin == "true" {
				label, ok := v["label"]
				if !ok {
					label = strings.ToTitle(k)
				}
				a[k] = strings.Trim(label, "\"")
			}
		}
	}
	return a
}

func Load() (err error) {
	var dir string
	if cfgDir == "" {
		dir = os.Getenv("IMSTO_CONF")
		if dir == "" {
			err = errors.New("IMSTO_CONF not found in environment or -conf unset")
			return
		}
	} else {
		dir = cfgDir
	}
	cfgFile := path.Join(dir, "imsto.ini")

	loadedConfig, err = ini.LoadFile(cfgFile)

	if err != nil {
		log.Print(err)
		return
	} /*else {
		log.Print("config loaded from " + cfgFile)
	}*/

	for _, f := range afterCalles {
		err = f()
		if err != nil {
			log.Printf("loaded call error: %s %s", err, cfgFile)
		}
	}

	return
}

func SetLogFile(name string) error {
	if logDir != "" {
		logfile := path.Join(logDir, name+".log")
		// log.Printf("logfile: %s", logfile)
		fd, err := os.OpenFile(logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664)
		if err != nil {
			log.Printf("logfile %s create failed", logfile)
			return err
		}
		log.SetOutput(fd)
	} else {
		log.Print("log dir is empty")
	}
	return nil
}

var (
	afterCalles [](func() error)
)

func AtLoaded(f func() error) {
	afterCalles = append(afterCalles, f)
}
