package config

import (
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
max_file_size = 393216
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

	if confDir != "" {
		LoadConfig(confDir)
	}

}

func GetValue(section, name string) string {
	var (
		value string
		ok    bool
	)

	value, ok = loadedConfig.Get(section, name)
	if !ok {
		value, ok = defaultConfig.Get("common", name)
		if !ok {
			log.Printf("'%v' variable missing from '%v' section", name, section)
			return ""
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

// type option map[string]string

// func (opt option) Set(k, v string) {
// 	opt[k] = v
// }

// func (opt option) Get(k string) (v string) {
// 	return opt[k]
// }
