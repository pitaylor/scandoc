package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
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
		job.report(InProgress, "scanning")

		err := job.Scan()

		if err == nil {
			job.report(InProgress, "scanning done, queued for processing")
			s.PdfJobs <- job
		} else {
			job.report(InProgress, fmt.Sprintf("scan failed: %v", err))
		}
	}
}

func (s *Service) WorkPdfJobs() {
	for job := range s.PdfJobs {
		var err error
		stageGlob := "out*.tif"

		if job.Settings.Clean {
			job.report(InProgress, "cleaning images")
			err = job.CleanImages(stageGlob)
			stageGlob = "clean*.png"
		}

		if job.Settings.Pdf && err == nil {
			job.report(InProgress, "generating PDF")
			err = job.GeneratePDF(stageGlob)

			if err == nil {
				job.report(InProgress, "removing temporary files")
				err = job.CleanUp()
			}
		}

		if err == nil {
			job.report(Done, "done!")
		} else {
			job.report(Failed, fmt.Sprintf("failed: %v", err))
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

		settings := NewSettings()
		settings.ParseValues(req.Form)

		job := NewJob(s.Dir, req.Form.Get("Name"), settings)

		s.ScanJobs <- job

		job.report(InProgress, "queued for scanning")
	})

	http.HandleFunc("/ws", serveWs)

	http.Handle("/scans/", http.StripPrefix("/scans", http.FileServer(http.Dir(s.Dir))))

	log.Println("listening on port 8090")
	err = http.ListenAndServe(":8090", nil)

	if err != nil {
		log.Fatal(err)
	}
}
