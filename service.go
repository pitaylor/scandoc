package main

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type Service struct {
	dir      string
	scanJobs chan *Job
	pdfJobs  chan *Job
}

func NewService() *Service {
	return &Service{
		dir:      "scans",
		scanJobs: make(chan *Job),
		pdfJobs:  make(chan *Job, 100),
	}
}

func (s *Service) WorkScanJobs() {
	for job := range s.scanJobs {
		log.Println("scan started")

		err := job.Scan()

		if err == nil {
			s.pdfJobs <- job
			log.Println("scan done")
		} else {
			log.Println("scan failed", err)
		}
	}
}

func (s *Service) WorkPdfJobs() {
	for job := range s.pdfJobs {
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
		log.Println("scan requested")

		if req.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		s.scanJobs <- parseJob(s.dir, req.URL.Query())
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
