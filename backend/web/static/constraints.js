/* Preset constraints for evaluation policy (schema_version 1) */
window.NOEMA_PRESET_CONSTRAINTS = [
  {
    "id": "pii_exposure_risk",
    "category": "privacy",
    "classification_type": "severity",
    "description": "Assess whether the dataset contains personally identifiable information that could identify individuals directly or indirectly, including contextual linkage of names, addresses, contact details, or identifiable imagery.",
    "severity_levels": {
      "0": "No identifiable personal data detected",
      "1": "Limited or partially redacted personal identifiers present",
      "2": "Clear personally identifiable information or identity-linkable data present"
    }
  },
  {
    "id": "regulated_sensitive_data_presence",
    "category": "regulated_data",
    "classification_type": "severity",
    "description": "Determine whether the dataset contains regulated sensitive information such as medical records, financial account information, biometric identifiers, or government-issued identifiers.",
    "severity_levels": {
      "0": "No regulated sensitive data detected",
      "1": "Possible indirect references to sensitive regulated data",
      "2": "Explicit regulated sensitive information present"
    }
  },
  {
    "id": "data_provenance_or_consent_violation_risk",
    "category": "data_governance",
    "classification_type": "severity",
    "description": "Evaluate whether the dataset appears to contain content that may have been collected without proper user consent or originates from private communications or restricted sources.",
    "severity_levels": {
      "0": "No evidence of consent or provenance concerns",
      "1": "Ambiguous or uncertain provenance indicators",
      "2": "Strong indicators of private or non-consensual data collection"
    }
  },
  {
    "id": "safety_critical_advisory_presence",
    "category": "safety",
    "classification_type": "severity",
    "description": "Assess whether the dataset contains safety-critical guidance such as medical, legal, engineering, or operational instructions that could cause harm if followed incorrectly.",
    "severity_levels": {
      "0": "No safety-critical instructions present",
      "1": "General informational references without actionable steps",
      "2": "Actionable safety-critical procedural instructions present"
    }
  },
  {
    "id": "harm_enabling_content_risk",
    "category": "risk_assessment",
    "classification_type": "severity",
    "description": "Evaluate whether the dataset includes content that could enable harmful real-world actions if generated or amplified by AI systems, including operational misuse or harmful procedural knowledge.",
    "severity_levels": {
      "0": "No harm-enabling content detected",
      "1": "Contextual discussion of potentially harmful topics",
      "2": "Operational or actionable harm-enabling information present"
    }
  },
  {
    "id": "dataset_intended_use_mismatch",
    "category": "ai_governance",
    "classification_type": "severity",
    "description": "Determine whether the dataset content appears inconsistent with its stated intended use or declared labeling context, indicating potential governance or deployment risk.",
    "severity_levels": {
      "0": "Dataset appears aligned with intended use",
      "1": "Minor inconsistencies or ambiguous alignment",
      "2": "Clear mismatch between dataset content and intended use"
    }
  }
];
