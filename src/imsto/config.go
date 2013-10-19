package imsto

import (
	// "code.google.com/p/gcfg"
	"github.com/vaughan0/go-ini"
	"os"
	"path"
	"strings"
	// "sync"
	// "errors"
	"log"
)

const defaultConfigIni = `
meta_dsn = postgres://wp_content@localhost?sslmode=disable
meta_table = wp_storage.meta_wpitem
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

func initConfig() {
	defaultConfig, _ = ini.Load(strings.NewReader(defaultConfigIni))

	confDir = os.Getenv("IMSTO_CONF_DIR")
	if confDir == "" {
		confDir, _ = os.Getwd()
		// panic(errors.New("env IMSTO_CONF_DIR not found"))
	}

}

func getConfig(section, name string) string {
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

func loadConfig(dir string) error {
	confDir = dir
	cfgFile := path.Join(confDir, "imsto.ini")
	var err error

	loadedConfig, err = ini.LoadFile(cfgFile)

	if err != nil {
		log.Print(err)
	} else {
		log.Print("loaded " + cfgFile)
	}

	return err
}
