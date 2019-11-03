package config

import (
	"log"
	"os"

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

// Sections ...
type Sections map[string]Section

// Sizes ...
type Sizes []uint

// Has ...
func (z Sizes) Has(v uint) bool {
	for _, size := range z {
		if v == size {
			return true
		}
	}
	return false
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
	StageHost        string            `envconfig:"STAGE_HOST"`     // stage.example.org
	WatermarkFile    string            `envconfig:"WATERMARK_FILE"` // /opt/imsto/watermark.png
	WatermarkOpacity uint8             `envconfig:"WATERMARK_OPACITY" default:"30"`
	SupportSizes     Sizes             `envconfig:"SUPPORT_SIZE" default:"60,120,256"`
	Sections         map[string]string `envconfig:"SECTIONS"` // [roof]label
	Engines          map[string]string `envconfig:"ENGINES"`  // [roof]engine
	Buckets          map[string]string `envconfig:"BUCKETS"`  // [roof]bucket
}

// vars
var (
	Version = "dev"
	Name    = "imsto"
	cfgDir  string

	// Current ...
	Current = new(Config)
)

// InDevelop ...
func InDevelop() bool {
	return "dev" == Version
}

// Root ...
func Root() string {
	return cfgDir
}

func init() {
	if err := envconfig.Process(Name, Current); err != nil {
		log.Printf("envconfig init ERR %s", err)
	}
}

// GetSections return administrable sections
func GetSections() map[string]string {
	return Current.Sections
}

// GetEngine ...
func GetEngine(roof string) string {
	if v, ok := Current.Engines[roof]; ok {
		return v
	}
	return ""
}

// EnvOr ...
func EnvOr(key, dft string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return dft
}
