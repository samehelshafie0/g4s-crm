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

//go:embed pdf_extract.py
var pdfExtractor []byte

// ExtractedTable is one table found in a vendor PDF, as a matrix of cleaned
// cells. Mapping cells to fields is the caller's job, so the same column
// detection serves spreadsheets and PDFs alike.
type ExtractedTable struct {
	Page         int        `json:"page"`
	LineItemRows int        `json:"lineItemRows"`
	Rows         [][]string `json:"rows"`
}

// ExtractPDFTables runs the embedded extractor over a server-staged PDF in a
// private directory. The file is never executed and no user path reaches the
// command line beyond the staged copy.
func ExtractPDFTables(ctx context.Context, pdfPath string) ([]ExtractedTable, error) {
	directory := filepath.Dir(pdfPath)
	script := filepath.Join(directory, "extract.py")
	if err := os.WriteFile(script, pdfExtractor, 0600); err != nil {
		return nil, err
	}
	output := filepath.Join(directory, "tables.json")

	python := os.Getenv("PDF_PYTHON")
	if python == "" {
		python = "python3"
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, python, script, pdfPath, output)
	cmd.Dir = directory
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }

	result, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, errors.New("Reading this PDF took too long; try a smaller file")
		}
		var failure struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(result, &failure) == nil && failure.Error != "" {
			return nil, errors.New(failure.Error)
		}
		return nil, errors.New("PDF reading is unavailable; verify the server PDF dependencies")
	}

	body, err := os.ReadFile(output)
	if err != nil {
		return nil, errors.New("The PDF produced no readable tables")
	}
	var parsed struct {
		Tables []ExtractedTable `json:"tables"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return parsed.Tables, nil
}
