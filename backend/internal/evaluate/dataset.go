package evaluate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

	raw, err := io.ReadAll(src)
	if err != nil {
		return nil, Dataset{}, fmt.Errorf("could not read dataset")
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
	if err := json.Unmarshal(raw, &ds); err != nil {
		return nil, Dataset{}, fmt.Errorf("dataset must match schema")
	}
	if len(ds.Items) == 0 {
		return nil, Dataset{}, fmt.Errorf("dataset.items must be a non-empty array")
	}
	for i, item := range ds.Items {
		if item.ID == "" {
			return nil, Dataset{}, fmt.Errorf("dataset.items[%d].id is required", i)
		}
		if item.Text == "" {
			return nil, Dataset{}, fmt.Errorf("dataset.items[%d].text is required", i)
		}
	}
	return raw, ds, nil
}

func sampleDataset(ds Dataset, limit int) Dataset {
	if limit <= 0 || len(ds.Items) <= limit {
		return ds
	}
	return Dataset{Items: ds.Items[:limit]}
}
