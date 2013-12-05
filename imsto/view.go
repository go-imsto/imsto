package main

import (
	"encoding/json"
	"fmt"
	"wpst.me/calf/storage"
)

var cmdView = &Command{
	UsageLine: "view -s roof [-id ID]",
	Short:     "view a id for item or browse",
	Long: `
Just a test command
`,
}

var (
	vroof       string
	vid         string
	limit, skip int
)

func init() {
	cmdView.Run = runView
	cmdView.Flag.StringVar(&vid, "id", "", "entry id")
	cmdView.Flag.StringVar(&vroof, "s", "", "config section name")
	cmdView.Flag.IntVar(&skip, "skip", 0, "skip")
	cmdView.Flag.IntVar(&limit, "limit", 5, "limit")
}

func runView(args []string) bool {
	if vroof == "" {
		return false
	}
	if vid != "" {
		id, err := storage.NewEntryId(vid)
		if err != nil {
			fmt.Printf("error: %s", err)
			return true
		}
		// fmt.Println(id)

		var mw storage.MetaWrapper
		mw = storage.NewMetaWrapper(vroof)

		var entry *storage.Entry
		entry, err = mw.GetMeta(*id)

		if err != nil {
			fmt.Println(err)
			return true
		}

		bytes, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("found entry: %s\n", bytes)
		}

	} else {
		var mw storage.MetaWrapper
		mw = storage.NewMetaWrapper(vroof)

		a, t, err := mw.Browse(limit, skip, map[string]int{"created": storage.DESCENDING})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("total: %d\n", t)
		// fmt.Printf("rows: %s", a)
		for _, e := range a {
			fmt.Printf("entry %s %s %d %s\n", e.Id, e.Path, e.Size, e.Mime)
		}

	}

	return true
}
