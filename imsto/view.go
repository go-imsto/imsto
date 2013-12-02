package main

import (
	"fmt"
	"wpst.me/calf/storage"
)

var cmdView = &Command{
	UsageLine: "view ID",
	Short:     "view a id for item",
	Long: `
Just a test command
`,
}

func init() {
	cmdView.Run = runView
}

func runView(args []string) bool {
	al := len(args)
	if al == 0 {
		fmt.Println("noting")
	} else if args[0] == "browse" {
		var mw storage.MetaWrapper
		mw = storage.NewMetaWrapper("")
		limit := 5
		offset := 0
		a, t, err := mw.Browse(limit, offset, map[string]int{"created": -1})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("total: %d\n", t)
		// fmt.Printf("rows: %s", a)
		for _, e := range a {
			fmt.Printf("entry %s %s %d %s\n", e.Id, e.Path, e.Size, e.Mime)
		}
	} else {
		id, err := storage.NewEntryId(args[0])
		fmt.Println(id)

		var mw storage.MetaWrapper
		mw = storage.NewMetaWrapper("")

		var entry *storage.Entry
		entry, err = mw.GetMeta(*id)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("entry:", entry)

	}

	return true
}
