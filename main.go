package main

import (
	"flag"
)

func main() {
	s := NewService()

	flag.StringVar(&s.Dir, "dir", "scans", "directory for scans")
	flag.Parse()

	s.Start()
}
