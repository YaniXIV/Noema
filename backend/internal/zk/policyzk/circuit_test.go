package policyzk

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/test"
	"github.com/iden3/go-iden3-crypto/poseidon"
)

func TestPolicyGateCircuit_PoseidonCommitment(t *testing.T) {
	tOp := test.WithCurves(ecc.BN254)

	assert := test.NewAssert(t)

	enabled := [N]int{1, 0, 1, 0, 1, 1}
	maxAllowed := [N]int{0, 2, 1, 2, 0, 1}
	severity := [N]int{0, 2, 1, 0, 0, 1}

	overallPass, maxSev := computePolicyResult(enabled, maxAllowed, severity)

	commitment := computePoseidonCommitment(
		20260208,
		17, 23,
		enabled, maxAllowed, severity,
	)

	var circuit PolicyGateCircuit
	witness := PolicyGateCircuit{
		DatasetDigestLo: 17,
		DatasetDigestHi: 23,
		Commitment:      commitment,
		OverallPass:     overallPass,
		MaxSeverity:     maxSev,
	}
	for i := 0; i < N; i++ {
		witness.Enabled[i] = enabled[i]
		witness.MaxAllowed[i] = maxAllowed[i]
		witness.Severity[i] = severity[i]
	}

	assert.ProverSucceeded(&circuit, &witness)
}

func computePolicyResult(enabled [N]int, maxAllowed [N]int, severity [N]int) (int, int) {
	anyFail := false
	maxSev := 0
	for i := 0; i < N; i++ {
		if enabled[i] == 0 {
			continue
		}
		if severity[i] > maxSev {
			maxSev = severity[i]
		}
		if severity[i] > maxAllowed[i] {
			anyFail = true
		}
	}
	if anyFail {
		return 0, maxSev
	}
	return 1, maxSev
}

func computePoseidonCommitment(domainSep int, digestLo int, digestHi int, enabled [N]int, maxAllowed [N]int, severity [N]int) *big.Int {
	inputs := make([]*big.Int, 0, 3+3*N)
	inputs = append(inputs, big.NewInt(int64(domainSep)))
	inputs = append(inputs, big.NewInt(int64(digestLo)))
	inputs = append(inputs, big.NewInt(int64(digestHi)))
	for i := 0; i < N; i++ {
		inputs = append(inputs, big.NewInt(int64(enabled[i])))
	}
	for i := 0; i < N; i++ {
		inputs = append(inputs, big.NewInt(int64(maxAllowed[i])))
	}
	for i := 0; i < N; i++ {
		inputs = append(inputs, big.NewInt(int64(severity[i])))
	}
	return poseidonHashChunkss(inputs)
}

func poseidonHashChunkss(inputs []*big.Int) *big.Int {
	const maxInputs = 16
	if len(inputs) <= maxInputs {
		out, err := poseidon.Hash(inputs)
		if err != nil {
			panic(err)
		}
		return out
	}
	out1, err := poseidon.Hash(inputs[:maxInputs])
	if err != nil {
		panic(err)
	}
	rest := make([]*big.Int, 0, 1+len(inputs[maxInputs:]))
	rest = append(rest, out1)
	rest = append(rest, inputs[maxInputs:]...)
	out2, err := poseidon.Hash(rest)
	if err != nil {
		panic(err)
	}
	return out2
}
