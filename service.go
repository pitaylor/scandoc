package main

import (
	"embed"
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
		log.Println("job started")

		var err error
		stageGlob := "out*.tif"

		if job.Settings.Clean {
			err = job.CleanImages(stageGlob)
			stageGlob = "clean*.png"
		}

		if job.Settings.Pdf && err == nil {
			err = job.GeneratePDF(stageGlob)

			if err == nil {
				err = job.CleanUp()
			}
		}

		if err == nil {
			log.Println("job done")
		} else {
			log.Println("job failed", err)
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

		s.ScanJobs <- parseJob(s.Dir, req.Form)
	})

	log.Println("listening on port 8090")
	err = http.ListenAndServe(":8090", nil)

	if err != nil {
		log.Fatal(err)
	}
}

// parseJob creates a Job from POST/GET parameters.
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

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
