package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Job struct {
	Name     string
	Dir      string
	Settings *Settings
}

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
		Name:     path + suffix + ".pdf",
		Dir:      path + suffix,
		Settings: settings,
	}
}

// Scan scans a document using scanimage and produces a .tif file for each page in Dir.
func (j *Job) Scan() error {
	_, err := os.Stat(j.Dir)

	if os.IsNotExist(err) {
		err = os.MkdirAll(j.Dir, os.ModePerm)
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

	cmd.Dir = j.Dir

	err = runCommand(cmd)

	if err != nil {
		_ = j.CleanUp()
	}

	return err
}

// GeneratePDF creates a PDF named Name using img2pdf and ocrmypdf on the .tif files in Dir.
func (j *Job) GeneratePDF() error {
	files, err := filepath.Glob(filepath.Join(j.Dir, "out*.tif"))

	if err != nil {
		return err
	}

	sort.Slice(files, func(i, j int) bool {
		return parseIndex(files[i]) < parseIndex(files[j])
	})

	var args []string
	pdfFile := filepath.Join(j.Dir, "out.pdf")
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
		"--deskew",
		"--clean",
		"--clean-final",
		pdfFile,
		j.Name,
	)
	err = runCommand(cmd)

	if err != nil {
		return err
	}

	return j.CleanUp()
}

func (j *Job) CleanUp() error {
	return os.RemoveAll(j.Dir)
}

func runCommand(cmd *exec.Cmd) error {
	log.Println("command:", cmd)

	output, err := cmd.CombinedOutput()

	log.Printf(strings.TrimSuffix(string(output), "\n"))

	return err
}

func parseIndex(path string) int {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	str := strings.TrimPrefix(base, "out")
	i, err := strconv.Atoi(str)

	if err != nil {
		log.Println("error parsing batch index", err)
	}

	return i
}
