package evaluate

import (
	"encoding/json"
	"fmt"
	"mime/multipart"

	"noema/internal/config"
)

func parseSpec(form *multipart.Form) (Spec, error) {
	specStrs := form.Value["spec"]
	if len(specStrs) == 0 || specStrs[0] == "" {
		return Spec{}, fmt.Errorf("missing field: spec")
	}
	var spec Spec
	if err := json.Unmarshal([]byte(specStrs[0]), &spec); err != nil {
		return Spec{}, fmt.Errorf("invalid spec JSON")
	}
	return spec, nil
}

func validateSpec(spec Spec) error {
	if spec.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema_version")
	}
	for _, cn := range spec.Constraints {
		if !ValidateAllowedMaxSeverity(cn.AllowedMaxSeverity) {
			return fmt.Errorf("constraint allowed_max_severity must be 0, 1, or 2")
		}
	}
	for _, cn := range spec.CustomConstraints {
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
	datasetFile := datasetFiles[0]
	if datasetFile.Size > config.MaxDatasetBytes {
		return nil, nil, fmt.Errorf("dataset exceeds 50MB limit")
	}
	if err := validateDatasetJSON(datasetFile); err != nil {
		return nil, nil, err
	}

	imageFiles := form.File["images"]
	if len(imageFiles) > config.MaxImages {
		return nil, nil, fmt.Errorf("maximum 10 images allowed")
	}
	for _, f := range imageFiles {
		if f.Size > config.MaxImageBytes {
			return nil, nil, fmt.Errorf("each image must be at most 5MB")
		}
	}
	return datasetFile, imageFiles, nil
}

func validateDatasetJSON(fh *multipart.FileHeader) error {
	_, _, err := readDatasetFile(fh)
	return err
}
