package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"time"
)

//go:embed ui/build
var embeddedFiles embed.FS

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
		job.report(StatusInProgress, "scanning")

		err := job.Scan()

		if err == nil {
			job.report(StatusInProgress, "scanning done, queued for processing")
			s.PdfJobs <- job
		} else {
			job.report(StatusInProgress, fmt.Sprintf("scan failed: %v", err))
		}
	}
}

func (s *Service) WorkPdfJobs() {
	for job := range s.PdfJobs {
		var err error
		stageGlob := "out*.tif"

		if job.settings.Clean {
			job.report(StatusInProgress, "cleaning images")
			err = job.CleanImages(stageGlob)
			stageGlob = "clean*.png"
		}

		if job.settings.Pdf && err == nil {
			job.report(StatusInProgress, "generating PDF")
			err = job.GeneratePDF(stageGlob)

			if err == nil {
				job.report(StatusInProgress, "removing temporary files")
				err = job.CleanUp()
			}
		}

		if err == nil {
			job.report(StatusDone, "done!")
		} else {
			job.report(StatusFailed, fmt.Sprintf("failed: %v", err))
		}
	}
}

func (s *Service) Start() {
	go s.WorkScanJobs()
	go s.WorkPdfJobs()

	fileSystem, err := fs.Sub(embeddedFiles, "ui/build")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(fileSystem))

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			fileServer.ServeHTTP(w, req)
			return
		}

		err := req.ParseForm()

		if err != nil {
			log.Println("error parsing request")
		}

		log.Println("scan requested", req.Form)

		if req.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		job := s.parseJob(req.Form)
		job.report(StatusInProgress, "queued for scanning")
		s.ScanJobs <- job
	})

	http.HandleFunc("/ws", serveWs)

	log.Println("listening on port 8090")
	err = http.ListenAndServe(":8090", nil)

	if err != nil {
		log.Fatal(err)
	}
}

// parseJob creates a Job from POST/GET parameters.
func (s *Service) parseJob(query url.Values) *Job {
	name := time.Now().Format("2006-01-02")

	if query.Has("name") {
		name += " " + query.Get("name")
	} else {
		name += " Document"
	}

	settings := NewSettings()
	settings.ParseValues(query)

	return NewJob(s.Dir, name, settings)
}
