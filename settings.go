package main

import (
	"log"
	"net/url"
	"strconv"
)

type Settings struct {
	Source     string `json:"source"`
	Mode       string `json:"mode"`
	Resolution int    `json:"resolution"`
	Brightness int    `json:"brightness"`
	Contrast   int    `json:"contrast"`
	Clean      bool   `json:"clean"`
	Pdf        bool   `json:"pdf"`
}

func NewSettings() *Settings {
	return &Settings{
		Source:     "ADF Front",
		Mode:       "Gray",
		Resolution: 300,
		Brightness: 0,
		Contrast:   0,
		Clean:      true,
		Pdf:        true,
	}
}

func (s *Settings) ParseValues(values url.Values) {
	for k, _ := range values {
		stringVal := values.Get(k)
		switch k {
		case "source":
			s.Source = stringVal
		case "mode":
			s.Mode = stringVal
		case "resolution":
			if value, err := strconv.Atoi(stringVal); err == nil {
				s.Resolution = value
			} else {
				log.Printf("error parsing %v: %v\n", k, err)
			}
		case "brightness":
			if value, err := strconv.Atoi(stringVal); err == nil {
				s.Brightness = value
			} else {
				log.Printf("error parsing %v: %v\n", k, err)
			}
		case "contrast":
			if value, err := strconv.Atoi(stringVal); err == nil {
				s.Contrast = value
			} else {
				log.Printf("error parsing %v: %v\n", k, err)
			}
		case "clean":
			if stringVal == "false" {
				s.Clean = false
			}
		case "pdf":
			if stringVal == "false" {
				s.Pdf = false
			}
		}
	}
}
