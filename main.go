package main

import (
	"github.com/go-imsto/imsto/cmd"
	_ "github.com/go-imsto/imsto/storage/backend/file"
	_ "github.com/go-imsto/imsto/storage/backend/s3c"
)

func main() {
	cmd.Main()
}
