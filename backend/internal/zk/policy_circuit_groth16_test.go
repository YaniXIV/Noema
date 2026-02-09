package zk

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"

	"noema/internal/zk/policyzk"
)

func TestPolicyGateCircuit_Groth16PassFail(t *testing.T) {
	var circuit policyzk.PolicyGateCircuit
	r1cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		t.Fatalf("compile circuit: %v", err)
	}

	pk, vk, err := groth16.Setup(r1cs)
	if err != nil {
		t.Fatalf("setup groth16: %v", err)
	}

	datasetDigestHex := "00112233445566778899aabbccddeeffffeeddccbbaa99887766554433221100"
	enabled := [PolicyGateConstraintCount]uint64{1, 1, 1, 0, 1, 0}
	maxAllowed := [PolicyGateConstraintCount]uint64{1, 2, 0, 1, 2, 0}

	passSeverity := [PolicyGateConstraintCount]uint64{1, 2, 0, 2, 1, 2}
	passCommitment, err := CommitmentPoseidon(datasetDigestHex, enabled, maxAllowed, passSeverity)
	if err != nil {
		t.Fatalf("commitment pass: %v", err)
	}
	passCommitmentInt, err := parseCommitmentHex(passCommitment)
	if err != nil {
		t.Fatalf("parse commitment pass: %v", err)
	}
	passOverall := 1
	passMaxSeverity := maxSeverity(enabled, passSeverity)

	passLo, passHi, err := datasetDigestLimbs(datasetDigestHex)
	if err != nil {
		t.Fatalf("dataset limbs: %v", err)
	}

	passAssignment := policyzk.PolicyGateCircuit{
		DatasetDigestLo: passLo,
		DatasetDigestHi: passHi,
		Enabled:         toVarArray(enabled),
		MaxAllowed:      toVarArray(maxAllowed),
		Severity:        toVarArray(passSeverity),
		Commitment:      passCommitmentInt,
		OverallPass:     passOverall,
		MaxSeverity:     passMaxSeverity,
	}

	passWitness, err := frontend.NewWitness(&passAssignment, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatalf("pass witness: %v", err)
	}
	passProof, err := groth16.Prove(r1cs, pk, passWitness)
	if err != nil {
		t.Fatalf("pass prove: %v", err)
	}
	passPublic, err := passWitness.Public()
	if err != nil {
		t.Fatalf("pass public: %v", err)
	}
	if err := groth16.Verify(passProof, vk, passPublic); err != nil {
		t.Fatalf("pass verify: %v", err)
	}

	failSeverity := [PolicyGateConstraintCount]uint64{1, 2, 1, 2, 1, 2}
	failCommitment, err := CommitmentPoseidon(datasetDigestHex, enabled, maxAllowed, failSeverity)
	if err != nil {
		t.Fatalf("commitment fail: %v", err)
	}
	failCommitmentInt, err := parseCommitmentHex(failCommitment)
	if err != nil {
		t.Fatalf("parse commitment fail: %v", err)
	}
	failOverall := 0
	failMaxSeverity := maxSeverity(enabled, failSeverity)

	failAssignment := policyzk.PolicyGateCircuit{
		DatasetDigestLo: passLo,
		DatasetDigestHi: passHi,
		Enabled:         toVarArray(enabled),
		MaxAllowed:      toVarArray(maxAllowed),
		Severity:        toVarArray(failSeverity),
		Commitment:      failCommitmentInt,
		OverallPass:     failOverall,
		MaxSeverity:     failMaxSeverity,
	}

	failWitness, err := frontend.NewWitness(&failAssignment, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatalf("fail witness: %v", err)
	}
	failProof, err := groth16.Prove(r1cs, pk, failWitness)
	if err != nil {
		t.Fatalf("fail prove: %v", err)
	}
	failPublic, err := failWitness.Public()
	if err != nil {
		t.Fatalf("fail public: %v", err)
	}
	if err := groth16.Verify(failProof, vk, failPublic); err != nil {
		t.Fatalf("fail verify: %v", err)
	}

	badAssignment := failAssignment
	badAssignment.OverallPass = 1
	badWitness, err := frontend.NewWitness(&badAssignment, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatalf("bad witness: %v", err)
	}
	if _, err := groth16.Prove(r1cs, pk, badWitness); err == nil {
		t.Fatalf("expected proving to fail with incorrect OverallPass")
	}
}

func maxSeverity(enabled, severity [PolicyGateConstraintCount]uint64) int {
	has1 := false
	for i := 0; i < PolicyGateConstraintCount; i++ {
		if enabled[i] == 0 {
			continue
		}
		if severity[i] >= 2 {
			return 2
		}
		if severity[i] == 1 {
			has1 = true
		}
	}
	if has1 {
		return 1
	}
	return 0
}
