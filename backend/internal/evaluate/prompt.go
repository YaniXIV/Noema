package evaluate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const promptVersion = "noema-eval-v1"

type PromptConstraint struct {
	ID                 string
	Description        string
	SeverityLevels     map[string]string
	AllowedMaxSeverity int
}

var presetConstraints = map[string]PromptConstraint{
	"pii_exposure_risk": {
		ID:          "pii_exposure_risk",
		Description: "Assess whether the dataset contains personally identifiable information that could identify individuals directly or indirectly, including contextual linkage of names, addresses, contact details, or identifiable imagery.",
		SeverityLevels: map[string]string{
			"0": "No identifiable personal data detected",
			"1": "Limited or partially redacted personal identifiers present",
			"2": "Clear personally identifiable information or identity-linkable data present",
		},
	},
	"regulated_sensitive_data_presence": {
		ID:          "regulated_sensitive_data_presence",
		Description: "Determine whether the dataset contains regulated sensitive information such as medical records, financial account information, biometric identifiers, or government-issued identifiers.",
		SeverityLevels: map[string]string{
			"0": "No regulated sensitive data detected",
			"1": "Possible indirect references to sensitive regulated data",
			"2": "Explicit regulated sensitive information present",
		},
	},
	"data_provenance_or_consent_violation_risk": {
		ID:          "data_provenance_or_consent_violation_risk",
		Description: "Evaluate whether the dataset appears to contain content that may have been collected without proper user consent or originates from private communications or restricted sources.",
		SeverityLevels: map[string]string{
			"0": "No evidence of consent or provenance concerns",
			"1": "Ambiguous or uncertain provenance indicators",
			"2": "Strong indicators of private or non-consensual data collection",
		},
	},
	"safety_critical_advisory_presence": {
		ID:          "safety_critical_advisory_presence",
		Description: "Assess whether the dataset contains safety-critical guidance such as medical, legal, engineering, or operational instructions that could cause harm if followed incorrectly.",
		SeverityLevels: map[string]string{
			"0": "No safety-critical instructions present",
			"1": "General informational references without actionable steps",
			"2": "Actionable safety-critical procedural instructions present",
		},
	},
	"harm_enabling_content_risk": {
		ID:          "harm_enabling_content_risk",
		Description: "Evaluate whether the dataset includes content that could enable harmful real-world actions if generated or amplified by AI systems, including operational misuse or harmful procedural knowledge.",
		SeverityLevels: map[string]string{
			"0": "No harm-enabling content detected",
			"1": "Contextual discussion of potentially harmful topics",
			"2": "Operational or actionable harm-enabling information present",
		},
	},
	"dataset_intended_use_mismatch": {
		ID:          "dataset_intended_use_mismatch",
		Description: "Determine whether the dataset content appears inconsistent with its stated intended use or declared labeling context, indicating potential governance or deployment risk.",
		SeverityLevels: map[string]string{
			"0": "Dataset appears aligned with intended use",
			"1": "Minor inconsistencies or ambiguous alignment",
			"2": "Clear mismatch between dataset content and intended use",
		},
	},
}

func buildSystemPrompt() string {
	return strings.TrimSpace(`
You are a strict dataset policy evaluator for Noema.
Return ONLY valid JSON that conforms to the provided schema.
Be concise: each rationale should be a short phrase or sentence (<= 25 words).
Do not include any extra keys, markdown, or explanations outside JSON.
	`)
}

func buildUserPrompt(spec Spec, datasetJSON []byte, images []ImageInfo) string {
	constraints := resolvePromptConstraints(spec)
	var buf bytes.Buffer
	buf.WriteString("Evaluate the dataset against the following constraints.\n")
	buf.WriteString("Constraints:\n")
	for _, c := range constraints {
		buf.WriteString(fmt.Sprintf("- id: %s\n", c.ID))
		buf.WriteString(fmt.Sprintf("  description: %s\n", c.Description))
		if len(c.SeverityLevels) > 0 {
			buf.WriteString("  severity_levels:\n")
			for _, k := range sortedSeverityKeys(c.SeverityLevels) {
				buf.WriteString(fmt.Sprintf("    %s: %s\n", k, c.SeverityLevels[k]))
			}
		}
		buf.WriteString(fmt.Sprintf("  allowed_max_severity: %d\n", c.AllowedMaxSeverity))
	}
	if len(images) > 0 {
		buf.WriteString("Images attached (matched by items[].image_ref to filename):\n")
		for _, img := range images {
			buf.WriteString(fmt.Sprintf("- %s (%s)\n", img.Filename, img.MIMEType))
		}
	}
	buf.WriteString("Dataset JSON (possibly sampled):\n")
	buf.Write(datasetJSON)
	return buf.String()
}

func resolvePromptConstraints(spec Spec) []PromptConstraint {
	var out []PromptConstraint
	for _, c := range spec.Constraints {
		if !c.Enabled {
			continue
		}
		if preset, ok := presetConstraints[c.ID]; ok {
			preset.AllowedMaxSeverity = c.AllowedMaxSeverity
			out = append(out, preset)
			continue
		}
		out = append(out, PromptConstraint{
			ID:                 c.ID,
			Description:        "No description provided for this constraint.",
			AllowedMaxSeverity: c.AllowedMaxSeverity,
		})
	}
	for _, c := range spec.CustomConstraints {
		if !c.Enabled {
			continue
		}
		desc := strings.TrimSpace(c.Description)
		if desc == "" {
			desc = strings.TrimSpace(c.Title)
		}
		if desc == "" {
			desc = "Custom constraint."
		}
		out = append(out, PromptConstraint{
			ID:                 c.ID,
			Description:        desc,
			AllowedMaxSeverity: c.AllowedMaxSeverity,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func sortedSeverityKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func marshalSampledDataset(ds Dataset) ([]byte, error) {
	return json.MarshalIndent(ds, "", "  ")
}
