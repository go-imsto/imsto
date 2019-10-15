package main

import (
	"github.com/go-imsto/imsto/cmd"
	_ "github.com/go-imsto/imsto/storage/backend/file"
	_ "github.com/go-imsto/imsto/storage/backend/grid"
	_ "github.com/go-imsto/imsto/storage/backend/qiniu"
	_ "github.com/go-imsto/imsto/storage/backend/s3"
)

func main() {
	cmd.Main()
}
