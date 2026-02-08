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

func TestDecodePublicInputsValidation(t *testing.T) {
	_, err := DecodePublicInputs([]byte("noema_public_inputs_v1|pt=1|ms=1|op=1|c=not-hex"))
	if err == nil {
		t.Fatalf("expected validation error for commitment format")
	}

	_, err = DecodePublicInputs([]byte("noema_public_inputs_v1|pt=3|ms=1|op=1|c=0xabc123"))
	if err == nil {
		t.Fatalf("expected validation error for policy threshold range")
	}

	_, err = DecodePublicInputs([]byte("noema_public_inputs_v1|pt=1|ms=1|c=0xabc123"))
	if err == nil {
		t.Fatalf("expected validation error for missing fields")
	}

	_, err = DecodePublicInputs([]byte("noema_public_inputs_v1|pt=1|pt=2|ms=1|op=1|c=0xabc123"))
	if err == nil {
		t.Fatalf("expected validation error for duplicate fields")
	}
}
