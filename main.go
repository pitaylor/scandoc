package main

import (
	"flag"
)

func main() {
	s := NewService()

	flag.StringVar(&s.dir, "dir", "scans", "directory for scans")
	flag.Parse()

	s.Start()
}
