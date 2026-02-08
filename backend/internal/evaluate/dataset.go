package evaluate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"noema/internal/config"
)

type Dataset struct {
	Items []DatasetItem `json:"items"`
}

type DatasetItem struct {
	ID       string         `json:"id"`
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata,omitempty"`
	ImageRef string         `json:"image_ref,omitempty"`
}

func readDatasetFile(fh *multipart.FileHeader) ([]byte, Dataset, error) {
	src, err := fh.Open()
	if err != nil {
		return nil, Dataset{}, fmt.Errorf("could not read dataset")
	}
	defer src.Close()

	raw, err := io.ReadAll(io.LimitReader(src, int64(config.MaxDatasetBytes)+1))
	if err != nil {
		return nil, Dataset{}, fmt.Errorf("could not read dataset")
	}
	if len(raw) > config.MaxDatasetBytes {
		return nil, Dataset{}, fmt.Errorf("dataset exceeds limit of %s", formatBytes(int64(config.MaxDatasetBytes)))
	}
	if len(raw) == 0 {
		return nil, Dataset{}, fmt.Errorf("dataset file is empty")
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, Dataset{}, fmt.Errorf("dataset must be valid JSON")
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, Dataset{}, fmt.Errorf("dataset must be a single JSON value")
	}

	var ds Dataset
	dec = json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&ds); err != nil {
		return nil, Dataset{}, fmt.Errorf("dataset must match schema")
	}
	if len(ds.Items) == 0 {
		return nil, Dataset{}, fmt.Errorf("dataset.items must be a non-empty array")
	}
	seenIDs := make(map[string]struct{}, len(ds.Items))
	for i, item := range ds.Items {
		trimmedID := strings.TrimSpace(item.ID)
		if trimmedID == "" {
			return nil, Dataset{}, fmt.Errorf("dataset.items[%d].id is required", i)
		}
		if trimmedID != item.ID {
			return nil, Dataset{}, fmt.Errorf("dataset.items[%d].id must not include leading/trailing whitespace", i)
		}
		if strings.TrimSpace(item.Text) == "" {
			return nil, Dataset{}, fmt.Errorf("dataset.items[%d].text is required", i)
		}
		if item.ImageRef != "" {
			trimmedRef := strings.TrimSpace(item.ImageRef)
			if trimmedRef == "" {
				return nil, Dataset{}, fmt.Errorf("dataset.items[%d].image_ref must be non-empty", i)
			}
			if trimmedRef != item.ImageRef {
				return nil, Dataset{}, fmt.Errorf("dataset.items[%d].image_ref must not include leading/trailing whitespace", i)
			}
			if filepath.Base(item.ImageRef) != item.ImageRef {
				return nil, Dataset{}, fmt.Errorf("dataset.items[%d].image_ref must not include path separators", i)
			}
		}
		if _, exists := seenIDs[item.ID]; exists {
			return nil, Dataset{}, fmt.Errorf("dataset.items[%d].id must be unique", i)
		}
		seenIDs[item.ID] = struct{}{}
	}
	return raw, ds, nil
}

func sampleDataset(ds Dataset, limit int) Dataset {
	if limit <= 0 || len(ds.Items) <= limit {
		return ds
	}
	return Dataset{Items: ds.Items[:limit]}
}
