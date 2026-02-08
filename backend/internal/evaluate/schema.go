package evaluate

func evalResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []any{
			"schema_version",
			"constraints",
			"max_severity",
		},
		"properties": map[string]any{
			"schema_version": map[string]any{
				"type": "integer",
				"enum": []any{1},
			},
			"constraints": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items": map[string]any{
					"type": "object",
					"required": []any{
						"id",
						"severity",
						"rationale",
					},
					"properties": map[string]any{
						"id": map[string]any{
							"type": "string",
						},
						"severity": map[string]any{
							"type":    "integer",
							"minimum": 0,
							"maximum": 2,
						},
						"rationale": map[string]any{
							"type": "string",
						},
					},
					"additionalProperties": false,
				},
			},
			"max_severity": map[string]any{
				"type":    "integer",
				"minimum": 0,
				"maximum": 2,
			},
			"confidence": map[string]any{
				"type":    "number",
				"minimum": 0,
				"maximum": 1,
			},
		},
		"additionalProperties": false,
	}
}
