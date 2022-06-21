package main

import (
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type JobStatus int

type Job struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Dir      string    `json:"dir"`
	Status   JobStatus `json:"status"`
	Message  string    `json:"message"`
	Settings *Settings `json:"settings"`
	Client   *Client   `json:"-"`
}

const (
	InProgress JobStatus = iota
	Done
	Failed
)

var numberRegex = regexp.MustCompile("[0-9]+")

func (t JobStatus) String() string {
	return [...]string{"in_progress", "done", "failed"}[t]
}

func (t JobStatus) MarshalText() (text []byte, err error) {
	return []byte(t.String()), nil
}

func NewJob(dir string, name string, settings *Settings) *Job {
	if name == "" {
		name = time.Now().Format("2006-01-02") + " Document"
	}

	i := 0
	path := filepath.Join(dir, name)
	suffix := ""

	for {
		if i > 0 {
			suffix = " " + strconv.Itoa(i+1)
		}

		_, err1 := os.Stat(path + suffix)
		_, err2 := os.Stat(path + suffix + ".pdf")

		if os.IsNotExist(err1) && os.IsNotExist(err2) {
			break
		}

		i += 1
	}

	return &Job{
		Id:       uuid.NewString(),
		Name:     name + suffix,
		Dir:      dir,
		Settings: settings,
	}
}

func (j *Job) dir() string {
	return filepath.Join(j.Dir, j.Name)
}

func (j *Job) url() string {
	url := "/scans/" + j.Name
	if j.Settings.Pdf {
		url += ".pdf"
	}
	return url
}

// Scan scans a document using scanimage and produces a .tif file for each page in Dir named `outN.tif`.
func (j *Job) Scan() error {
	_, err := os.Stat(j.dir())

	if os.IsNotExist(err) {
		err = os.MkdirAll(j.dir(), os.ModePerm)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command(
		"scanimage",
		"--format=tiff",
		"--batch",
		"--source", j.Settings.Source,
		"--mode", j.Settings.Mode,
		"--resolution", strconv.Itoa(j.Settings.Resolution),
		"--brightness", strconv.Itoa(j.Settings.Brightness),
		"--contrast", strconv.Itoa(j.Settings.Contrast),
		"--page-height", "0",
	)

	cmd.Dir = j.dir()

	err = runCommand(cmd)

	if err != nil {
		_ = j.CleanUp()
	}

	return err
}

// CleanImages cleans up scanned images specified by `globPattern` using NoteShrink and produces files named
// `cleanN.png` in Dir.
func (j *Job) CleanImages(globPattern string) error {
	files, err := j.globFiles(globPattern)

	if err != nil {
		return err
	}

	var args []string
	args = append(args, "-c", "true") // skip pdf conversion
	args = append(args, "-b", filepath.Join(j.dir(), "clean"))
	args = append(args, files...)

	cmd := exec.Command("noteshrink", args...)
	return runCommand(cmd)
}

// GeneratePDF creates a PDF named `Name` from image files specified by `globPattern` using img2pdf and ocrmypdf.
func (j *Job) GeneratePDF(globPattern string) error {
	files, err := j.globFiles(globPattern)

	if err != nil {
		return err
	}

	var args []string
	pdfFile := filepath.Join(j.dir(), "out.pdf")
	args = append(args, "--output", pdfFile)
	args = append(args, files...)

	cmd := exec.Command("img2pdf", args...)
	err = runCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command(
		"ocrmypdf",
		"--rotate-pages",
		"--clean",
		pdfFile,
		filepath.Join(j.Dir, j.Name)+".pdf",
	)
	return runCommand(cmd)
}

func (j *Job) CleanUp() error {
	return os.RemoveAll(j.dir())
}

func (j *Job) globFiles(globPattern string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(j.dir(), globPattern))

	if err != nil {
		return files, err
	}

	sort.Slice(files, func(i, j int) bool {
		return parseIndex(files[i]) < parseIndex(files[j])
	})

	return files, nil
}

func (j *Job) report(status JobStatus, message string) {
	log.Printf("job report: %v - %v\n", status, message)

	j.Status = status
	j.Message = message

	if j.Client != nil {
		j.Client.queueResponse(j)
	}
}

func runCommand(cmd *exec.Cmd) error {
	log.Println("command:", cmd)

	output, err := cmd.CombinedOutput()

	log.Printf(strings.TrimSuffix(string(output), "\n"))

	return err
}

func parseIndex(path string) int {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	matches := numberRegex.FindAllString(base, 1)

	i := 0
	var err error

	if len(matches) != 0 {
		i, err = strconv.Atoi(matches[0])

		if err != nil {
			log.Println("error parsing batch index", err)
		}
	}

	return i
}
