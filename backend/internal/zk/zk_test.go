package zk

import (
	"encoding/base64"
	"testing"
)

func TestProofRoundTrip(t *testing.T) {
	witness := testWitnessInputs()
	commitment, err := CommitmentPoseidon(witness.DatasetDigestHex, witness.Enabled, witness.MaxAllowed, witness.Severity)
	if err != nil {
		t.Fatalf("CommitmentPoseidon error: %v", err)
	}
	pi := PublicInputs{
		PolicyThreshold: 0,
		MaxSeverity:     2,
		OverallPass:     true,
		Commitment:      commitment,
		Witness:         witness,
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

func TestProofMismatch(t *testing.T) {
	witness := testWitnessInputs()
	commitment, err := CommitmentPoseidon(witness.DatasetDigestHex, witness.Enabled, witness.MaxAllowed, witness.Severity)
	if err != nil {
		t.Fatalf("CommitmentPoseidon error: %v", err)
	}
	pi := PublicInputs{
		PolicyThreshold: 0,
		MaxSeverity:     2,
		OverallPass:     true,
		Commitment:      commitment,
		Witness:         witness,
	}
	proof, err := GenerateProof(pi)
	if err != nil {
		t.Fatalf("GenerateProof error: %v", err)
	}

	badPub, err := EncodePublicInputs(PublicInputs{
		PolicyThreshold: 0,
		MaxSeverity:     2,
		OverallPass:     false,
		Commitment:      commitment,
	})
	if err != nil {
		t.Fatalf("EncodePublicInputs error: %v", err)
	}
	ok, _, err := VerifyProof(proof.ProofB64, base64.StdEncoding.EncodeToString(badPub))
	if err != nil {
		t.Fatalf("VerifyProof error: %v", err)
	}
	if ok {
		t.Fatalf("expected proof to fail")
	}
}

func testWitnessInputs() *WitnessInputs {
	return &WitnessInputs{
		DatasetDigestHex: "00112233445566778899aabbccddeeffffeeddccbbaa99887766554433221100",
		Enabled:          [PolicyGateConstraintCount]uint64{1, 1, 1, 0, 1, 0},
		MaxAllowed:       [PolicyGateConstraintCount]uint64{1, 2, 0, 1, 2, 0},
		Severity:         [PolicyGateConstraintCount]uint64{1, 2, 0, 2, 1, 2},
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
