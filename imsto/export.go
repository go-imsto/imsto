package main

import (
	"fmt"
	// "log"
	"os"
	"path"
	// "strings"
	"wpst.me/calf/storage"
)

var cmdExport = &Command{
	UsageLine: "export -s roof -o direcotry",
	Short:     "export data to local file",
	Long: `
export to local system
`,
}

var (
	eroof  string
	edir   string
	eid    string
	etotal int
	elimit int
	eskip  int
)

const (
	max_limit = 50
)

func init() {
	cmdExport.Run = runExport
	cmdExport.Flag.StringVar(&eroof, "s", "", "config section name")
	cmdExport.Flag.StringVar(&edir, "o", "", "a local direcotry to export into.")
	cmdExport.Flag.StringVar(&eid, "id", "", "only export a special id.")
	cmdExport.Flag.IntVar(&etotal, "total", 0, "export total count.")
	cmdExport.Flag.IntVar(&elimit, "limit", 10, "page size.")
	cmdExport.Flag.IntVar(&eskip, "skip", 0, "offset.")
}

func runExport(args []string) bool {

	if eroof == "" || edir == "" {
		return false
	}

	mw := storage.NewMetaWrapper(eroof)
	if eid != "" {
		id, err := storage.NewEntryId(eid)
		if err != nil {
			fmt.Printf("error id: %s, %s", eid, err)
			return false
		}
		entry, err := mw.GetEntry(*id)
		if err != nil {
			fmt.Printf("get entry error: %s", err)
			return false
		}
		return _save_export(entry, edir)
	}

	filter := storage.MetaFilter{}

	total, err := mw.Count(filter)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Printf("total: %d\n", total)

	if total == 0 {
		return true
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
			return false
		}

		for _, entry := range a {
			if !_save_export(entry, edir) {
				return false
			}
		}
		skip += limit
	}

	return true
}

func _save_export(entry *storage.Entry, edir string) bool {
	name := path.Join(edir, entry.Path)
	fmt.Printf("save to: %s ", name)
	if fi, fe := os.Stat(name); fe == nil && fi.Size() == int64(entry.Size) {
		fmt.Println("exist")
		return true
	}
	err := storage.Dump(entry, eroof, name)
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		fmt.Print("ok\n")
	}
	return true
}
