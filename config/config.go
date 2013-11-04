package config

import (
	// "code.google.com/p/gcfg"
	"github.com/vaughan0/go-ini"
	"os"
	"path"
	"strings"
	// "sync"
	// "errors"
	// "flag"
	"log"
)

const defaultConfigIni = `
meta_dsn = postgres://imsto@localhost?sslmode=disable
meta_table = imsto.meta_wpitem
engine = mongodb
db_name = storage
thumb_path = /thumb
thumb_root = /opt/imsto/cache/thumb/
tmp_dir = /tmp/
support_size = 120,160,400
`

// var once sync.Once

var (
	confDir       string
	defaultConfig ini.File
	loadedConfig  ini.File
)

func GetConfDir() string {
	return confDir
}

func SetConfDir(dir string) {

	if _, err := os.Stat(dir); err != nil {
		log.Println(err)
	}

	confDir = dir
}

func init() {
	defaultConfig, _ = ini.Load(strings.NewReader(defaultConfigIni))

	confDir = os.Getenv("IMSTO_CONF_DIR")
	if confDir == "" {
		log.Println("env IMSTO_CONF_DIR not found")
		confDir, _ = os.Getwd()
		// panic(errors.New("env IMSTO_CONF_DIR not found"))
	}

	LoadConfig(confDir)
}

func GetValue(section, name string) string {
	var (
		value string
		ok    bool
	)

	value, ok = loadedConfig.Get(section, name)
	if !ok {
		value, ok = defaultConfig.Get("", name)
		if !ok {
			log.Printf("'%v' variable missing from '%v' section", name, section)
			return ""
		}
	}

	return value
}

func LoadConfig(dir string) error {
	cfgFile := path.Join(dir, "imsto.ini")
	var err error

	loadedConfig, err = ini.LoadFile(cfgFile)

	if err != nil {
		log.Print(err)
	} else {
		log.Print("loaded " + cfgFile)
	}

	return err
}

type option map[string]string

func (opt option) Set(k, v string) {
	opt[k] = v
}

func (opt option) Get(k string) (v string) {
	return opt[k]
}
