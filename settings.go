package main

import (
	"log"
	"net/url"
	"strconv"
)

type Settings struct {
	Source     string
	Mode       string
	Resolution int
	Brightness int
	Contrast   int
}

func NewSettings() *Settings {
	return &Settings{
		Source:     "ADF Front",
		Mode:       "Gray",
		Resolution: 300,
		Brightness: 0,
		Contrast:   0,
	}
}

func (s *Settings) ParseValues(values url.Values) {
	for k, _ := range values {
		switch k {
		case "source":
			s.Source = values.Get(k)
		case "mode":
			s.Mode = values.Get(k)
		case "resolution":
			if value, err := strconv.Atoi(values.Get(k)); err != nil {
				s.Resolution = value
			} else {
				log.Printf("error parsing %v: %v\n", k, err)
			}
		case "brightness":
			if value, err := strconv.Atoi(values.Get(k)); err != nil {
				s.Brightness = value
			} else {
				log.Printf("error parsing %v: %v\n", k, err)
			}
		case "contrast":
			if value, err := strconv.Atoi(values.Get(k)); err != nil {
				s.Contrast = value
			} else {
				log.Printf("error parsing %v: %v\n", k, err)
			}
		}
	}
}
