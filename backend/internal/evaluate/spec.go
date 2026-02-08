package evaluate

// Spec is the parsed evaluation spec (schema_version 1).
type Spec struct {
	SchemaVersion     int                `json:"schema_version"`
	EvaluationName    string             `json:"evaluation_name"`
	Policy            Policy             `json:"policy"`
	Constraints       []Constraint       `json:"constraints"`
	CustomConstraints []CustomConstraint `json:"custom_constraints"`
}

type Policy struct {
	Reveal RevealPolicy `json:"reveal"`
}

type RevealPolicy struct {
	MaxSeverity bool `json:"max_severity"`
	Commitment  bool `json:"commitment"`
}

type Constraint struct {
	ID                 string `json:"id"`
	Enabled            bool   `json:"enabled"`
	AllowedMaxSeverity int    `json:"allowed_max_severity"` // 0, 1, or 2
}

type CustomConstraint struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	Enabled            bool   `json:"enabled"`
	AllowedMaxSeverity int    `json:"allowed_max_severity"`
}

// ValidateAllowedMaxSeverity returns true if v is 0, 1, or 2.
func ValidateAllowedMaxSeverity(v int) bool {
	return v >= 0 && v <= 2
}
