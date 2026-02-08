package evaluate

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
)

var runIDCounter uint64

func genRunID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1e9))
	if err == nil {
		return fmt.Sprintf("run_%d_%d", time.Now().UnixMilli(), n.Int64())
	}
	counter := atomic.AddUint64(&runIDCounter, 1)
	return fmt.Sprintf("run_%d_%d_%d", time.Now().UnixMilli(), os.Getpid(), counter)
}

func ensureRunDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func saveRunFiles(runPath string, dataset *multipart.FileHeader, images []*multipart.FileHeader) error {
	if err := saveUpload(dataset, filepath.Join(runPath, "dataset.json")); err != nil {
		return fmt.Errorf("failed to save dataset: %w", err)
	}
	for i, f := range images {
		ext := filepath.Ext(f.Filename)
		if ext == "" {
			ext = ".bin"
		}
		dst := filepath.Join(runPath, fmt.Sprintf("image_%d%s", i, ext))
		if err := saveUpload(f, dst); err != nil {
			return fmt.Errorf("failed to save image %d: %w", i, err)
		}
	}
	return nil
}

func saveUpload(fh *multipart.FileHeader, dst string) error {
	src, err := fh.Open()
	if err != nil {
		return fmt.Errorf("open upload: %w", err)
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	if err != nil {
		return fmt.Errorf("copy to %s: %w", dst, err)
	}
	return nil
}

func saveJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

type RunIndexEntry struct {
	RunID          string `json:"run_id"`
	Status         string `json:"status"`
	Timestamp      int64  `json:"ts"`
	EvaluationName string `json:"evaluation_name,omitempty"`
}

func updateRunsIndex(runsDir string, limit int, entry RunIndexEntry) error {
	if limit <= 0 {
		return nil
	}
	indexPath := filepath.Join(runsDir, "index.json")
	var entries []RunIndexEntry
	var corruptedErr error
	if b, err := os.ReadFile(indexPath); err == nil {
		if err := json.Unmarshal(b, &entries); err != nil {
			backup := indexPath + ".corrupt-" + strconv.FormatInt(time.Now().UnixMilli(), 10)
			if renameErr := os.Rename(indexPath, backup); renameErr != nil {
				corruptedErr = fmt.Errorf("runs index corrupted; failed to archive: %w", renameErr)
			} else {
				corruptedErr = fmt.Errorf("runs index corrupted; archived as %s", backup)
			}
		}
	}
	entries = append([]RunIndexEntry{entry}, entries...)
	if len(entries) > limit {
		entries = entries[:limit]
	}
	if err := saveJSON(indexPath, entries); err != nil {
		return err
	}
	return corruptedErr
}
