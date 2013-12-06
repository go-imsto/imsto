package main

import (
	"fmt"
	// "log"
	// "os"
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
	etotal int
)

const (
	max_limit = 50
)

func init() {
	cmdExport.Run = runExport
	cmdExport.Flag.StringVar(&eroof, "s", "", "config section name")
	cmdExport.Flag.StringVar(&edir, "o", "", "a local direcotry to export into.")
	cmdExport.Flag.IntVar(&etotal, "total", 0, "export total count.")
	cmdExport.Flag.IntVar(&elimit, "limit", 10, "export total count.")
}

func runExport(args []string) bool {

	if eroof == "" || edir == "" {
		return false
	}

	mw := storage.NewMetaWrapper(eroof)
	total, err := mw.Count()
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
		skip  = 0
	)
	if total < max_limit {
		limit = total
	}

	for skip < total {
		fmt.Printf("start %d/%d\n", skip, total)
		a, err := mw.Browse(limit, skip, map[string]int{"created": storage.DESCENDING})
		if err != nil {
			fmt.Println(err)
			return false
		}

		for _, entry := range a {
			data, err := storage.FetchBlob(entry, eroof)
			if err != nil {
				fmt.Println(err)
				return false
			}
			filename := path.Join(edir, entry.Path)
			fmt.Printf("save to: %s ", filename)
			err = storage.SaveFile(filename, data)
			if err != nil {
				fmt.Println(err)
				return false
			} else {
				fmt.Print("ok\n")
			}
		}
		skip += limit
	}

	return true
}
