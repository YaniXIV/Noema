package zk

import (
	"encoding/base64"
	"strings"
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

// --- NEW TESTS ---

func TestCannotClaimPassWhenWitnessFailsPolicy(t *testing.T) {
	// Make a witness that FAILS the policy (enabled=1 and severity>maxAllowed).
	witness := testWitnessInputs()
	witness.Enabled[0] = 1
	witness.MaxAllowed[0] = 0
	witness.Severity[0] = 2 // 2 > 0 => fail when enabled

	commitment, err := CommitmentPoseidon(witness.DatasetDigestHex, witness.Enabled, witness.MaxAllowed, witness.Severity)
	if err != nil {
		t.Fatalf("CommitmentPoseidon error: %v", err)
	}

	// Try to LIE publicly: claim OverallPass=true even though witness violates constraints.
	pi := PublicInputs{
		PolicyThreshold: 0,
		MaxSeverity:     2,    // doesn't matter; OverallPass is already wrong
		OverallPass:     true, // <- lie
		Commitment:      commitment,
		Witness:         witness,
	}

	// If the circuit is correct, proving should fail (unsatisfied constraints).
	// If your GenerateProof doesn't surface that error, then verification MUST fail.
	proof, err := GenerateProof(pi)
	if err == nil {
		ok, _, verr := VerifyProof(proof.ProofB64, proof.PublicInputsB64)
		if verr != nil {
			t.Fatalf("VerifyProof error (unexpected): %v", verr)
		}
		if ok {
			t.Fatalf("expected proof to NOT verify when claiming pass with failing witness (circuit may not be enforcing OverallPass correctly)")
		}
	}
}

func TestFailingWitnessCanProveWithOverallPassFalse(t *testing.T) {
	// Same failing witness as above, but now we claim the correct public result: OverallPass=false.
	// This SHOULD be provable if your circuit is meant to *output* pass/fail rather than require pass.
	witness := testWitnessInputs()
	witness.Enabled[0] = 1
	witness.MaxAllowed[0] = 0
	witness.Severity[0] = 2 // 2 > 0 => fail

	commitment, err := CommitmentPoseidon(witness.DatasetDigestHex, witness.Enabled, witness.MaxAllowed, witness.Severity)
	if err != nil {
		t.Fatalf("CommitmentPoseidon error: %v", err)
	}

	pi := PublicInputs{
		PolicyThreshold: 0,
		MaxSeverity:     2,     // max severity among enabled should be 2 (we enabled index 0 w/ severity 2)
		OverallPass:     false, // correct
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
		t.Fatalf("expected proof to verify when failing witness is paired with OverallPass=false")
	}
}

func TestCommitmentMismatchFailsVerification(t *testing.T) {
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

	// Flip one hex nibble of the commitment in the PUBLIC INPUTS.
	badCommit := flipHexNibble(commitment)

	badPub, err := EncodePublicInputs(PublicInputs{
		PolicyThreshold: 0,
		MaxSeverity:     2,
		OverallPass:     true,
		Commitment:      badCommit,
	})
	if err != nil {
		t.Fatalf("EncodePublicInputs error: %v", err)
	}

	ok, _, err := VerifyProof(proof.ProofB64, base64.StdEncoding.EncodeToString(badPub))
	if err != nil {
		t.Fatalf("VerifyProof error: %v", err)
	}
	if ok {
		t.Fatalf("expected proof to fail when Commitment is changed (commitment may not be properly bound in-circuit)")
	}
}

func TestMaxSeverityMismatchFailsVerification(t *testing.T) {
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

	// Wrong MaxSeverity (should be 2 for the test witness; try 1).
	badPub, err := EncodePublicInputs(PublicInputs{
		PolicyThreshold: 0,
		MaxSeverity:     1, // <- wrong
		OverallPass:     true,
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
		t.Fatalf("expected proof to fail when MaxSeverity is changed (MaxSeverity may not be properly constrained/bound)")
	}
}

// --- helpers ---

// flipHexNibble flips the last hex nibble of a 0x-prefixed hex string.
// This is enough to guarantee a different value while keeping formatting valid.
func flipHexNibble(hexStr string) string {
	s := strings.TrimSpace(hexStr)
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	if len(s) == 0 {
		return "0x1"
	}

	last := s[len(s)-1]
	var flipped byte
	switch last {
	case '0':
		flipped = '1'
	default:
		flipped = '0'
	}
	out := s[:len(s)-1] + string(flipped)
	return "0x" + out
}
