package main

import (
	"flag"
)

var service *Service

func main() {
	service = NewService()

	flag.StringVar(&service.Dir, "dir", "scans", "directory for scans")
	flag.Parse()

	service.Start()
}
