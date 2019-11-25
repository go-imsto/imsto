package config

import (
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Section ...
type Section struct { // example: {demo,file,Demo,/var/lib/imsto/,demo.imsto.org}
	Name   string `json:"name,omitempty"`
	Engine string `json:"engine,omitempty"`
	Label  string `json:"label,omitempty"`
	Root   string `json:"root,omitempty"`
	Host   string `json:"host,omitempty"` // stage host
}

// Sizes ...
type Sizes []uint

// Has ...
func (z Sizes) Has(v uint) bool {
	for _, size := range z {
		if v == size || v/2 == size {
			return true
		}
	}
	return false
}

// IPNet ...
type IPNet struct{ net.IPNet }

// Decode ...
func (z *IPNet) Decode(value string) error {
	_, ipn, err := net.ParseCIDR(value)
	if err != nil {
		return err
	}
	*z = IPNet{*ipn}
	return nil
}

// Config ...
type Config struct {
	MaxFileSize      uint              `envconfig:"MAX_FILESIZE" default:"2097152"` // 2MB
	MaxWidth         uint              `envconfig:"MAX_WIDTH" default:"1600"`
	MaxHeight        uint              `envconfig:"MAX_HEIGHT" default:"1600"`
	MinWidth         uint              `envconfig:"MIN_WIDTH" default:"50"`
	MinHeight        uint              `envconfig:"MIN_HEIGHT" default:"50"`
	MaxQuality       uint8             `envconfig:"MAX_QUALITY" default:"88"`
	CacheRoot        string            `envconfig:"CACHE_ROOT" default:"/opt/imsto/cache/"`
	LocalRoot        string            `envconfig:"LOCAL_ROOT" default:"/var/lib/imsto/"`
	StageHost        string            `envconfig:"STAGE_HOST"`     // stage.example.org
	WatermarkFile    string            `envconfig:"WATERMARK_FILE"` // /opt/imsto/watermark.png
	WatermarkOpacity uint8             `envconfig:"WATERMARK_OPACITY" default:"30"`
	SupportSizes     Sizes             `envconfig:"SUPPORT_SIZE" default:"60,120,256"`
	Roofs            []string          `envconfig:"ROOFS" default:"demo"` // roof1,roof2
	Engines          map[string]string `envconfig:"ENGINES"`              // [roof]engine
	Prefixes         map[string]string `envconfig:"PREFIXES"`             // [roof]prefix
	WhiteList        []IPNet           `envconfig:"WHITELIST"`
	ReadTimeout      time.Duration     `envconfig:"READ_TIMEOUT" default:"10s"`
	TiringListen     string            `envconfig:"TIRING_LISTEN" default:":8967"`
	StageListen      string            `envconfig:"STAGE_LISTEN" default:":8968"`
	RPCListen        string            `envconfig:"RPC_LISTEN" default:":8969"`
}

// vars
var (
	Version = "dev"
	Name    = "imsto"

	// Current ...
	Current = new(Config)
)

// InDevelop ...
func InDevelop() bool {
	return "dev" == Version
}

func init() {
	if err := envconfig.Process(Name, Current); err != nil {
		log.Printf("envconfig init ERR %s", err)
	}

	if len(Current.LocalRoot) < 2 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("userHomeDir err %s", err)
			return
		}
		Current.LocalRoot = path.Join(homeDir, Name)
	}

	if len(Current.CacheRoot) < 2 {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Printf("userCacheDir err %s", err)
			return
		}
		Current.CacheRoot = path.Join(cacheDir, Name)
	}
}

// HasSection ...
func HasSection(roof string) bool {
	if _, ok := Current.Engines[roof]; ok {
		return true
	}
	return false
}

// GetEngine ...
func GetEngine(roof string) string {
	if v, ok := Current.Engines[roof]; ok {
		return v
	}
	return ""
}

// GetPrefix ...
func GetPrefix(roof string) string {
	if s, ok := Current.Prefixes[roof]; ok && len(s) > 0 {
		return s
	}
	return roof
}

// EnvOr ...
func EnvOr(key, dft string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return dft
}
