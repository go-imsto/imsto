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
view a id for item or browse
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
			return false
		}
		// fmt.Println(id)

		var mw storage.MetaWrapper
		mw = storage.NewMetaWrapper(vroof)

		var entry *storage.Entry
		entry, err = mw.GetMeta(*id)

		if err != nil {
			fmt.Println(err)
			return false
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
		t, err := mw.Count()
		if err != nil {
			fmt.Println(err)
			return false
		}

		a, err := mw.Browse(limit, skip, map[string]int{"created": storage.DESCENDING})
		if err != nil {
			fmt.Println(err)
			return false
		}

		fmt.Printf("total: %d\n", t)
		if t == 0 {
			fmt.Println("empty result")
			return true
		}
		fmt.Printf(" %26s %34s %9s %11s %13s\n", "id", "path", "size", "mime", "name")
		for _, e := range a {
			fmt.Printf(" %29s %35s %7d %11s %13s\n", e.Id, e.Path, e.Size, e.Mime, e.Name)
		}

	}

	return true
}
