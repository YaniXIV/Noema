package evaluate

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"noema/internal/config"
)

func parseSpec(form *multipart.Form) (Spec, error) {
	specStrs := form.Value["spec"]
	if len(specStrs) == 0 || specStrs[0] == "" {
		return Spec{}, fmt.Errorf("missing field: spec")
	}
	if len(specStrs) > 1 {
		return Spec{}, fmt.Errorf("only one spec value allowed")
	}
	var spec Spec
	dec := json.NewDecoder(strings.NewReader(specStrs[0]))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&spec); err != nil {
		return Spec{}, fmt.Errorf("invalid spec JSON")
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return Spec{}, fmt.Errorf("invalid spec JSON")
	}
	return spec, nil
}

func validateSpec(spec Spec) error {
	if spec.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema_version")
	}
	for _, cn := range spec.Constraints {
		if cn.Enabled && strings.TrimSpace(cn.ID) == "" {
			return fmt.Errorf("constraint id must be non-empty")
		}
		if !ValidateAllowedMaxSeverity(cn.AllowedMaxSeverity) {
			return fmt.Errorf("constraint allowed_max_severity must be 0, 1, or 2")
		}
	}
	for _, cn := range spec.CustomConstraints {
		if cn.Enabled && strings.TrimSpace(cn.ID) == "" {
			return fmt.Errorf("custom_constraint id must be non-empty")
		}
		if !ValidateAllowedMaxSeverity(cn.AllowedMaxSeverity) {
			return fmt.Errorf("custom_constraint allowed_max_severity must be 0, 1, or 2")
		}
	}
	return nil
}

func parseUploads(form *multipart.Form) (*multipart.FileHeader, []*multipart.FileHeader, error) {
	datasetFiles := form.File["dataset"]
	if len(datasetFiles) == 0 {
		return nil, nil, fmt.Errorf("missing required file: dataset")
	}
	if len(datasetFiles) > 1 {
		return nil, nil, fmt.Errorf("only one dataset file allowed")
	}
	datasetFile := datasetFiles[0]
	if datasetFile.Size > config.MaxDatasetBytes {
		return nil, nil, fmt.Errorf("dataset exceeds limit of %s", formatBytes(int64(config.MaxDatasetBytes)))
	}
	imageFiles := form.File["images"]
	if len(imageFiles) > config.MaxImages {
		return nil, nil, fmt.Errorf("maximum %d images allowed", config.MaxImages)
	}
	seenImageNames := make(map[string]struct{}, len(imageFiles))
	for _, f := range imageFiles {
		if _, exists := seenImageNames[f.Filename]; exists {
			return nil, nil, fmt.Errorf("image filenames must be unique")
		}
		seenImageNames[f.Filename] = struct{}{}
		if f.Size > config.MaxImageBytes {
			return nil, nil, fmt.Errorf("each image must be at most %s", formatBytes(int64(config.MaxImageBytes)))
		}
	}
	if err := validateDatasetJSON(datasetFile, imageFiles); err != nil {
		return nil, nil, err
	}
	return datasetFile, imageFiles, nil
}

func validateDatasetJSON(fh *multipart.FileHeader, imageFiles []*multipart.FileHeader) error {
	_, ds, err := readDatasetFile(fh)
	if err != nil {
		return err
	}
	if len(imageFiles) == 0 {
		for i, item := range ds.Items {
			if item.ImageRef != "" {
				return fmt.Errorf("dataset.items[%d].image_ref provided but no images uploaded", i)
			}
		}
		return nil
	}
	imageNames := make(map[string]struct{}, len(imageFiles))
	for _, img := range imageFiles {
		imageNames[img.Filename] = struct{}{}
	}
	for i, item := range ds.Items {
		if item.ImageRef == "" {
			continue
		}
		if _, ok := imageNames[item.ImageRef]; !ok {
			return fmt.Errorf("dataset.items[%d].image_ref must match an uploaded filename", i)
		}
	}
	return nil
}

func formatBytes(n int64) string {
	const kb = 1024
	const mb = 1024 * 1024
	switch {
	case n%mb == 0:
		return fmt.Sprintf("%dMB", n/mb)
	case n%kb == 0:
		return fmt.Sprintf("%dKB", n/kb)
	default:
		return fmt.Sprintf("%d bytes", n)
	}
}
