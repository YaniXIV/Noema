package evaluate

func evalResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []any{
			"eval_version",
			"results",
		},
		"properties": map[string]any{
			"eval_version": map[string]any{
				"type": "string",
				"enum": []any{"noema_eval_v1"},
			},
			"results": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items": map[string]any{
					"type": "object",
					"required": []any{
						"id",
						"severity",
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
						"confidence": map[string]any{
							"type":    "number",
							"minimum": 0,
							"maximum": 1,
						},
					},
					"additionalProperties": false,
				},
			},
		},
		"additionalProperties": false,
	}
}
