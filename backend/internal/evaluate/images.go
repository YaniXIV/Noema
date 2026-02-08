package evaluate

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"

	"noema/internal/config"
)

type ImageInfo struct {
	Filename string
	MIMEType string
	Data     []byte
}

func readImages(files []*multipart.FileHeader) ([]ImageInfo, error) {
	out := make([]ImageInfo, 0, len(files))
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			return nil, fmt.Errorf("could not read image %q: %w", fh.Filename, err)
		}
		data, err := io.ReadAll(io.LimitReader(src, int64(config.MaxImageBytes)+1))
		closeErr := src.Close()
		if err != nil {
			return nil, fmt.Errorf("could not read image %q: %w", fh.Filename, err)
		}
		if len(data) > config.MaxImageBytes {
			return nil, fmt.Errorf("image %q exceeds limit of %s", fh.Filename, formatBytes(int64(config.MaxImageBytes)))
		}
		if closeErr != nil {
			return nil, fmt.Errorf("could not close image %q: %w", fh.Filename, closeErr)
		}
		mimeType := strings.TrimSpace(fh.Header.Get("Content-Type"))
		if mimeType == "" {
			mimeType = mime.TypeByExtension(strings.ToLower(filepath.Ext(fh.Filename)))
		}
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		out = append(out, ImageInfo{
			Filename: fh.Filename,
			MIMEType: mimeType,
			Data:     data,
		})
	}
	return out, nil
}
