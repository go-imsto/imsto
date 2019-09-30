package cmd

import (
	"fmt"
	"github.com/go-imsto/imsto/storage"
	"log"
)

var cmdAuth = &Command{
	UsageLine: "auth [options]",
	Short:     "add a auth apiKey from the command-line",
	Long: `

`,
}

var (
	aname = cmdAuth.Flag.String("name", "", "a new app name")
	aver  = cmdAuth.Flag.Uint("ver", 1, "change api version")
	asave = cmdAuth.Flag.Bool("save", false, "save to db or not")
)

func init() {
	cmdAuth.Run = authApp
}

func authApp(args []string) bool {
	if *aname != "" {
		app := storage.NewApp(*aname)
		app.Version = storage.VerID(*aver)
		fmt.Printf("new app: %s\nkey: %s\nsalt:%s\n", app.Name, app.ApiKey, app.ApiSalt)
		if *asave {
			err := app.Save()
			if err != nil {
				log.Print(err)
			} else {
				fmt.Println("save ok")
			}
		}
		return true
	}

	return false
}
