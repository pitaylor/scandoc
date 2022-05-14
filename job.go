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
)

type Job struct {
	id       string
	name     string
	dir      string
	settings *Settings
}

var numberRegex = regexp.MustCompile("[0-9]+")

func NewJob(dir string, baseName string, settings *Settings) *Job {
	i := 0
	path := filepath.Join(dir, baseName)
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
		id:       uuid.NewString(),
		name:     path + suffix + ".pdf",
		dir:      path + suffix,
		settings: settings,
	}
}

// Scan scans a document using scanimage and produces a .tif file for each page in dir named `outN.tif`.
func (j *Job) Scan() error {
	_, err := os.Stat(j.dir)

	if os.IsNotExist(err) {
		err = os.MkdirAll(j.dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command(
		"scanimage",
		"--format=tiff",
		"--batch",
		"--source", j.settings.Source,
		"--mode", j.settings.Mode,
		"--resolution", strconv.Itoa(j.settings.Resolution),
		"--brightness", strconv.Itoa(j.settings.Brightness),
		"--contrast", strconv.Itoa(j.settings.Contrast),
		"--page-height", "0",
	)

	cmd.Dir = j.dir

	err = runCommand(cmd)

	if err != nil {
		_ = j.CleanUp()
	}

	return err
}

// CleanImages cleans up scanned images specified by `globPattern` using NoteShrink and produces files named
// `cleanN.png` in dir.
func (j *Job) CleanImages(globPattern string) error {
	files, err := j.globFiles(globPattern)

	if err != nil {
		return err
	}

	var args []string
	args = append(args, "-c", "true") // skip pdf conversion
	args = append(args, "-b", filepath.Join(j.dir, "clean"))
	args = append(args, files...)

	cmd := exec.Command("noteshrink", args...)
	return runCommand(cmd)
}

// GeneratePDF creates a PDF named `name` from image files specified by `globPattern` using img2pdf and ocrmypdf.
func (j *Job) GeneratePDF(globPattern string) error {
	files, err := j.globFiles(globPattern)

	if err != nil {
		return err
	}

	var args []string
	pdfFile := filepath.Join(j.dir, "out.pdf")
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
		j.name,
	)
	return runCommand(cmd)
}

func (j *Job) CleanUp() error {
	return os.RemoveAll(j.dir)
}

func (j *Job) globFiles(globPattern string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(j.dir, globPattern))

	if err != nil {
		return files, err
	}

	sort.Slice(files, func(i, j int) bool {
		return parseIndex(files[i]) < parseIndex(files[j])
	})

	return files, nil
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
