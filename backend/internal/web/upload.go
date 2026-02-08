package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"noema/internal/config"
	"noema/internal/httputil"

	"github.com/gin-gonic/gin"
)

// UploadData is passed to the upload template.
type UploadData struct {
	Error    string
	Success  bool
	FileName string
	FileSize string
}

// UploadGet renders the upload page (GET /upload).
func UploadGet(c *gin.Context, tmpl string, data UploadData) {
	c.HTML(http.StatusOK, tmpl, data)
}

// UploadPost handles POST /upload (multipart form "file"). Saves to uploadsDir with a generated filename.
func UploadPost(c *gin.Context, uploadTmpl string, uploadsDir string) {
	const multipartOverhead = 1 << 20 // allow small overhead for multipart headers
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxUploadBytes+multipartOverhead)
	form, err := c.MultipartForm()
	if err != nil {
		if httputil.IsBodyTooLarge(err) {
			UploadGet(c, uploadTmpl, UploadData{Error: "File exceeds 50MB limit."})
			return
		}
		UploadGet(c, uploadTmpl, UploadData{Error: "Invalid form."})
		return
	}
	defer form.RemoveAll()
	files := form.File["file"]
	if len(files) == 0 {
		UploadGet(c, uploadTmpl, UploadData{Error: "No file selected."})
		return
	}
	file := files[0]
	if file.Size > config.MaxUploadBytes {
		UploadGet(c, uploadTmpl, UploadData{Error: "File exceeds 50MB limit."})
		return
	}
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		UploadGet(c, uploadTmpl, UploadData{Error: "Server error saving file."})
		return
	}
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".bin"
	}
	base := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	safe := strings.Map(func(r rune) rune {
		if r == '.' || r == '/' || r == '\\' || r == 0 {
			return -1
		}
		return r
	}, base)
	if safe == "" {
		safe = "upload"
	}
	name := fmt.Sprintf("%s_%d%s", safe, time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadsDir, name)
	src, err := file.Open()
	if err != nil {
		UploadGet(c, uploadTmpl, UploadData{Error: "Could not read file."})
		return
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		UploadGet(c, uploadTmpl, UploadData{Error: "Server error saving file."})
		return
	}
	defer out.Close()
	written, err := io.Copy(out, src)
	if err != nil {
		os.Remove(dst)
		UploadGet(c, uploadTmpl, UploadData{Error: "Server error saving file."})
		return
	}
	if err := out.Sync(); err != nil {
		os.Remove(dst)
		UploadGet(c, uploadTmpl, UploadData{Error: "Server error saving file."})
		return
	}
	if file.Size > 0 && written != file.Size {
		os.Remove(dst)
		UploadGet(c, uploadTmpl, UploadData{Error: "Upload incomplete. Please try again."})
		return
	}
	sizeStr := formatSize(file.Size)
	UploadGet(c, uploadTmpl, UploadData{Success: true, FileName: name, FileSize: sizeStr})
}

func formatSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
