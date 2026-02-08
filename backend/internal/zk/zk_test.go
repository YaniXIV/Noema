package zk

import "testing"

func TestStubProofRoundTrip(t *testing.T) {
	pi := PublicInputs{
		PolicyThreshold: 1,
		MaxSeverity:     1,
		OverallPass:     true,
		Commitment:      "0xabc123",
	}
	proof, err := GenerateProof(pi)
	if err != nil {
		t.Fatalf("GenerateProof error: %v", err)
	}
	ok, _, err := VerifyProof(proof.ProofB64, proof.PublicInputsB64)
	if err != nil {
		t.Fatalf("VerifyProof error: %v", err)
	}
	if !ok {
		t.Fatalf("expected proof to verify")
	}
}

func TestStubProofMismatch(t *testing.T) {
	pi := PublicInputs{
		PolicyThreshold: 2,
		MaxSeverity:     0,
		OverallPass:     true,
		Commitment:      "0xdeadbeef",
	}
	proof, err := GenerateProof(pi)
	if err != nil {
		t.Fatalf("GenerateProof error: %v", err)
	}
	// Tamper with public inputs
	ok, _, err := VerifyProof(proof.ProofB64, "bm9lbWFfcHVibGljX2lucHV0c192MXxwdD0xfG1zPTB8b3A9MXxjPTB4ZGVhZGJlZWY=")
	if err != nil {
		t.Fatalf("VerifyProof error: %v", err)
	}
	if ok {
		t.Fatalf("expected proof to fail")
	}
}

func TestEncodePublicInputsValidation(t *testing.T) {
	_, err := EncodePublicInputs(PublicInputs{
		PolicyThreshold: 3,
		MaxSeverity:     0,
		OverallPass:     true,
		Commitment:      "0xabc",
	})
	if err == nil {
		t.Fatalf("expected validation error for policy threshold")
	}

	_, err = EncodePublicInputs(PublicInputs{
		PolicyThreshold: 1,
		MaxSeverity:     3,
		OverallPass:     true,
		Commitment:      "0xabc",
	})
	if err == nil {
		t.Fatalf("expected validation error for max severity")
	}

	_, err = EncodePublicInputs(PublicInputs{
		PolicyThreshold: 1,
		MaxSeverity:     1,
		OverallPass:     true,
		Commitment:      "",
	})
	if err == nil {
		t.Fatalf("expected validation error for commitment")
	}
}
