package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	// "strings"
	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage"
)

const usage_line = `export -s roof -o direcotry
`

const short_desc = "export data to local file"

var (
	cfgDir string
	roof   string
	edir   string
	eid    string
	etotal int
	elimit int
	eskip  int
)

const (
	max_limit = 50
)

func usage() {
	fmt.Printf("Usage: \t%s\nDefault Usage:\n", usage_line)
	flag.PrintDefaults()
	fmt.Println("\nDescription:\n   " + short_desc + "\n")
}

func init() {
	flag.StringVar(&cfgDir, "conf", "/etc/imsto", "app conf dir")
	flag.StringVar(&roof, "s", "", "config section name")
	flag.StringVar(&edir, "o", "", "a local direcotry to export into.")
	flag.StringVar(&eid, "id", "", "only export a special id.")
	flag.IntVar(&etotal, "total", 0, "export total count.")
	flag.IntVar(&elimit, "limit", 10, "page size.")
	flag.IntVar(&eskip, "skip", 0, "offset.")

	flag.Parse()
	if cfgDir != "" {
		config.SetRoot(cfgDir)
	}

	config.AtLoaded(func() error {
		return config.SetLogFile("export")
	})

	err := config.Load()
	if err != nil {
		log.Print("config load error: ", err)
		os.Exit(1)
	}

}

func main() {
	// fmt.Printf("roof: %s, edir: %s\n", roof, edir)
	if roof == "" || config.Root() == "" || edir == "" {
		usage()
		return
	}

	if !config.HasSection(roof) {
		fmt.Printf("roof [%s] not found\n", roof)
		return
	}

	if roof == "" || edir == "" {
		return
	}

	mw := storage.NewMetaWrapper(roof)
	if eid != "" {
		entry, err := mw.GetMapping(eid)
		if err != nil {
			fmt.Printf("get entry error: %s", err)
			return
		}
		_save_export(entry.Path, entry.Size, edir)
	}

	filter := storage.MetaFilter{}

	total, err := mw.Count(filter)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("total: %d\n", total)

	if total == 0 {
		return
	}

	var (
		limit = elimit
		skip  = eskip
	)
	if total < max_limit {
		limit = total
	}

	for skip < total {
		fmt.Printf("start %d/%d\n", skip, total)
		a, err := mw.Browse(limit, skip, map[string]int{"created": storage.DESCENDING}, filter)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, entry := range a {
			if !_save_export(entry.Path, entry.Size, edir) {
				return
			}
		}
		skip += limit
	}

	return
}

func _save_export(key string, size uint32, edir string) bool {
	name := path.Join(edir, key)
	fmt.Printf("save to: %s ", name)
	if fi, fe := os.Stat(name); fe == nil && fi.Size() == int64(size) {
		fmt.Println("exist")
		return true
	}
	err := storage.Dump(key, roof, name)
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		fmt.Print("ok\n")
	}
	return true
}
