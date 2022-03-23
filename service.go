package main

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type Service struct {
	Dir      string
	ScanJobs chan *Job
	PdfJobs  chan *Job
}

func NewService() *Service {
	return &Service{
		Dir:      "scans",
		ScanJobs: make(chan *Job),
		PdfJobs:  make(chan *Job, 100),
	}
}

func (s *Service) WorkScanJobs() {
	for job := range s.ScanJobs {
		log.Println("scan started")

		err := job.Scan()

		if err == nil {
			s.PdfJobs <- job
			log.Println("scan done")
		} else {
			log.Println("scan failed", err)
		}
	}
}

func (s *Service) WorkPdfJobs() {
	for job := range s.PdfJobs {
		log.Println("pdf started")

		err := job.GeneratePDF()

		if err == nil {
			log.Println("pdf done")
		} else {
			log.Println("pdf failed", err)
		}
	}
}

func (s *Service) Start() {
	go s.WorkScanJobs()
	go s.WorkPdfJobs()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()

		if err != nil {
			log.Println("error parsing request")
		}

		log.Println("scan requested", req.Form)

		if req.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		s.ScanJobs <- parseJob(s.Dir, req.Form)
	})

	log.Println("listening on port 8090")
	err := http.ListenAndServe(":8090", nil)

	if err != nil {
		log.Fatal(err)
	}
}

func parseJob(dir string, query url.Values) *Job {
	name := time.Now().Format("2006-01-02")

	if query.Has("name") {
		name += " " + query.Get("name")
	} else {
		name += " Document"
	}

	settings := NewSettings()
	settings.ParseValues(query)

	return NewJob(dir, name, settings)
}
