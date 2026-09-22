package services

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

//go:embed pdf_renderer.py
var pdfRenderer []byte

// RenderQuotePDF runs in a private job directory with a fresh Office profile.
// The caller stages only files opened through the upload root, never user paths.
func RenderQuotePDF(ctx context.Context, directory string, job any) (string, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		return "", err
	}
	if err = os.WriteFile(filepath.Join(directory, "job.json"), payload, 0600); err != nil {
		return "", err
	}
	script := filepath.Join(directory, "render.py")
	if err = os.WriteFile(script, pdfRenderer, 0600); err != nil {
		return "", err
	}
	python := os.Getenv("PDF_PYTHON")
	if python == "" {
		python = "python3"
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	output := filepath.Join(directory, "quote.pdf")
	cmd := exec.CommandContext(ctx, python, script, filepath.Join(directory, "job.json"), output)
	cmd.Dir = directory
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	result, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", errors.New("PDF export timed out or was cancelled; try fewer documents")
		}
		var failure struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(result, &failure) == nil && failure.Error != "" {
			return "", errors.New(failure.Error)
		}
		return "", errors.New("PDF renderer is unavailable; verify the server PDF dependencies")
	}
	return output, nil
}
