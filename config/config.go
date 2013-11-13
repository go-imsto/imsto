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
meta_dsn = postgres://imsto@localhost/imsto?sslmode=disable
meta_table_suffix = demo
engine = s3
bucket_name = imsto-demo
max_quality = 88
max_file_size = 262114
thumb_path = /thumb
thumb_root = /opt/imsto/cache/thumb/
tmp_dir = /tmp/
support_size = 120,160,400
`

// var once sync.Once

var (
	appRoot       string
	defaultConfig ini.File
	loadedConfig  ini.File
)

func AppRoot() string {
	return appRoot
}

func SetAppRoot(dir string) {

	if _, err := os.Stat(dir); err != nil {
		log.Println(err)
		return
	}

	appRoot = dir
}

func init() {
	defaultConfig, _ = ini.Load(strings.NewReader(defaultConfigIni))
}

func GetValue(section, name string) string {
	var (
		value string
		ok    bool
	)

	if value, ok = loadedConfig.Get(section, name); !ok {
		if value, ok = loadedConfig.Get("common", name); !ok {
			if value, ok = defaultConfig.Get("common", name); !ok {
				log.Printf("'%v' variable missing from '%v' section", name, section)
				return ""
			}
		}
	}

	return value
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

func Sections() []string {
	a := []string{}
	for name, _ := range loadedConfig {
		if name != "common" {
			a = append(a, name)
		}
	}
	return a
}

func Load() (err error) {
	var dir string
	if appRoot == "" {
		dir = os.Getenv("IMSTO_APP_ROOT")
		if dir == "" {
			err = errors.New("IMSTO_APP_ROOT not found in environment")
			// log.Println("IMSTO_APP_ROOT not found in environment, or -c dir unset")
			// dir, _ = os.Getwd()
			// panic(errors.New("env IMSTO_APP_ROOT not found"))
			return
		}
	} else {
		dir = appRoot
	}
	cfgFile := path.Join(dir, "config", "imsto.ini")

	loadedConfig, err = ini.LoadFile(cfgFile)

	if err != nil {
		log.Print(err)
	} else {
		log.Print("config loaded from " + cfgFile)
	}

	return
}

// type option map[string]string

// func (opt option) Set(k, v string) {
// 	opt[k] = v
// }

// func (opt option) Get(k string) (v string) {
// 	return opt[k]
// }
