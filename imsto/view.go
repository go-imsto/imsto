package main

import (
	"calf/storage"
	"fmt"
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
