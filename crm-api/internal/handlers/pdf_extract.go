package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"g4s-crm/api/internal/services"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Extraction is CPU-bound, so only a couple run at once; the rest are asked to
// retry rather than queueing behind a slow document.
var extractJobs = make(chan struct{}, 2)

const maxExtractBytes = 25 << 20

type ExtractHandler struct{ db *gorm.DB }

func NewExtractHandler(db *gorm.DB) *ExtractHandler { return &ExtractHandler{db: db} }

// PDFTables reads the tables out of an uploaded vendor PDF. The file is parsed
// in a private temporary directory and deleted immediately; nothing is stored,
// so attaching the document to a record stays a separate, deliberate step.
func (h *ExtractHandler) PDFTables(c *gin.Context) {
	select {
	case extractJobs <- struct{}{}:
		defer func() { <-extractJobs }()
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "error": gin.H{"message": "PDF reading is busy. Try again shortly."}})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxExtractBytes+(1<<20))
	if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
		response.BadRequest(c, "Upload must be a multipart form of at most 25 MB")
		return
	}
	defer c.Request.MultipartForm.RemoveAll()

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File is required")
		return
	}
	defer file.Close()
	if !strings.EqualFold(filepath.Ext(header.Filename), ".pdf") {
		response.BadRequest(c, "Only PDF files are read here; spreadsheets are read in the browser")
		return
	}

	directory, err := os.MkdirTemp("", "crm-pdf-extract-")
	if err != nil {
		apiError(c, err)
		return
	}
	defer os.RemoveAll(directory)

	// The staged name is ours, never the uploaded one, so no caller-controlled
	// path or extension reaches the filesystem.
	staged := filepath.Join(directory, "source.pdf")
	dst, err := os.OpenFile(staged, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		apiError(c, err)
		return
	}
	written, copyErr := io.Copy(dst, io.LimitReader(file, maxExtractBytes+1))
	dst.Close()
	if copyErr != nil {
		apiError(c, copyErr)
		return
	}
	if written > maxExtractBytes {
		response.BadRequest(c, "PDF must be 25 MB or smaller")
		return
	}

	tables, err := services.ExtractPDFTables(c.Request.Context(), staged)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	response.OK(c, gin.H{"fileName": filepath.Base(header.Filename), "tables": tables})
}
