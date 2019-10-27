package cmd

import (
	"fmt"

	"github.com/go-imsto/imsto/storage"
)

var cmdFetch = &Command{
	UsageLine: "fetch -uri URI -roof ROOF",
	Short:     "fetch a image from URI",
	Long: `
fetch a image from URI
`,
}

var (
	fetchRoof  = cmdFetch.Flag.String("roof", "", "roof")
	fetchURI   = cmdFetch.Flag.String("uri", "", "A imgae uri")
	fetchRefer = cmdFetch.Flag.String("refer", "", "http referer")
)

func init() {
	cmdFetch.Run = runFetch
}

func runFetch(args []string) bool {
	if *fetchRoof == "" || *fetchURI == "" {
		return false
	}
	fmt.Printf("Fetching into %s from %s\n", *fetchRoof, *fetchURI)
	in := storage.FetchInput{URI: *fetchURI, Roof: *fetchRoof}
	if *fetchRefer != "" {
		in.Referer = *fetchRefer
	}
	entry, err := storage.Fetch(in)
	if err != nil {
		logger().Warnw("fetch fail", "err", err)
		return true
	}
	fmt.Println("new entry ", entry.Path)
	return true
}
