package main

import (
	"flag"
	"fmt"
	"hash/crc64"
	"io/ioutil"
	"log"
)

var (
	filename string
)

func init() {
	flag.StringVar(&filename, "f", "", "filename")
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	if filename == "" {
		flag.PrintDefaults()
		return
	}
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	l := len(data)
	if l == 0 {
		fmt.Printf("%d\t0\t%s\n", l, filename)
		return
	}
	var s uint64
	s = crc64.Checksum(data, crc64.MakeTable(crc64.ISO))
	fmt.Printf("%08d\t%016x\t%s\n", len(data), s, filename)
	s = crc64.Checksum(data, crc64.MakeTable(crc64.ECMA))
	fmt.Printf("%08d\t%016x\t%s\n", len(data), s, filename)
}
