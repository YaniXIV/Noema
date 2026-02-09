package policyzk

import (
	"math/big"
	"testing"

	"github.com/AlpinYukseloglu/poseidon-gnark/circuits"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/stretchr/testify/require"
)

func TestPolicyGateCircuit_Groth16PassFail(t *testing.T) {
	require := require.New(t)

	var circuit PolicyGateCircuit
	r1cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	require.NoError(err)

	pk, vk, err := groth16.Setup(r1cs)
	require.NoError(err)

	datasetLo := big.NewInt(123456789)
	datasetHi := big.NewInt(987654321)

	enabled := []uint64{1, 1, 1, 0, 1, 0}
	maxAllowed := []uint64{1, 2, 0, 1, 2, 0}

	passSeverity := []uint64{1, 2, 0, 2, 1, 2}
	passCommitment := commitmentForCase(datasetLo, datasetHi, enabled, maxAllowed, passSeverity)
	passAssignment := PolicyGateCircuit{
		DatasetDigestLo: datasetLo,
		DatasetDigestHi: datasetHi,
		Enabled:         toVarArray(enabled),
		MaxAllowed:      toVarArray(maxAllowed),
		Severity:        toVarArray(passSeverity),
		Commitment:      passCommitment,
		OverallPass:     1,
		MaxSeverity:     2,
	}

	fullPass, err := frontend.NewWitness(&passAssignment, ecc.BN254.ScalarField())
	require.NoError(err)
	passProof, err := groth16.Prove(r1cs, pk, fullPass)
	require.NoError(err)
	passPublic, err := fullPass.Public()
	require.NoError(err)
	require.NoError(groth16.Verify(passProof, vk, passPublic))

	failSeverity := []uint64{1, 2, 1, 2, 1, 2}
	failCommitment := commitmentForCase(datasetLo, datasetHi, enabled, maxAllowed, failSeverity)
	failAssignment := PolicyGateCircuit{
		DatasetDigestLo: datasetLo,
		DatasetDigestHi: datasetHi,
		Enabled:         toVarArray(enabled),
		MaxAllowed:      toVarArray(maxAllowed),
		Severity:        toVarArray(failSeverity),
		Commitment:      failCommitment,
		OverallPass:     0,
		MaxSeverity:     2,
	}

	fullFail, err := frontend.NewWitness(&failAssignment, ecc.BN254.ScalarField())
	require.NoError(err)
	failProof, err := groth16.Prove(r1cs, pk, fullFail)
	require.NoError(err)
	failPublic, err := fullFail.Public()
	require.NoError(err)
	require.NoError(groth16.Verify(failProof, vk, failPublic))

	// Same failing inputs but incorrect OverallPass should not satisfy constraints.
	badAssignment := failAssignment
	badAssignment.OverallPass = 1
	fullBad, err := frontend.NewWitness(&badAssignment, ecc.BN254.ScalarField())
	require.NoError(err)
	_, err = groth16.Prove(r1cs, pk, fullBad)
	require.Error(err)
}

func toVarArray(vals []uint64) [N]frontend.Variable {
	var out [N]frontend.Variable
	for i := 0; i < N; i++ {
		out[i] = vals[i]
	}
	return out
}

func commitmentForCase(datasetLo, datasetHi *big.Int, enabled, maxAllowed, severity []uint64) *big.Int {
	inputs := make([]*big.Int, 0, 3+3*N)
	inputs = append(inputs, big.NewInt(20260208))
	inputs = append(inputs, new(big.Int).Set(datasetLo))
	inputs = append(inputs, new(big.Int).Set(datasetHi))
	for i := 0; i < N; i++ {
		inputs = append(inputs, new(big.Int).SetUint64(enabled[i]))
	}
	for i := 0; i < N; i++ {
		inputs = append(inputs, new(big.Int).SetUint64(maxAllowed[i]))
	}
	for i := 0; i < N; i++ {
		inputs = append(inputs, new(big.Int).SetUint64(severity[i]))
	}
	return poseidonHashChunksNative(inputs)
}

func poseidonHashChunksNative(inputs []*big.Int) *big.Int {
	const maxInputs = 16
	if len(inputs) <= maxInputs {
		return poseidonNative(inputs)
	}
	h := poseidonNative(inputs[:maxInputs])
	rest := make([]*big.Int, 0, 1+len(inputs[maxInputs:]))
	rest = append(rest, h)
	rest = append(rest, inputs[maxInputs:]...)
	return poseidonNative(rest)
}

func poseidonNative(inputs []*big.Int) *big.Int {
	out := poseidonExNative(inputs, big.NewInt(0), 1)
	return out[0]
}

func poseidonExNative(inputs []*big.Int, initialState *big.Int, nOuts int) []*big.Int {
	t := len(inputs) + 1
	nRoundsPC := [16]int{56, 57, 56, 60, 60, 63, 64, 63, 60, 66, 60, 65, 70, 60, 64, 68}
	nRoundsF := 8
	nRoundsP := nRoundsPC[t-2]

	c := bigIntSliceToElements(circuits.POSEIDON_C(t))
	s := bigIntSliceToElements(circuits.POSEIDON_S(t))
	m := bigIntMatrixToElements(circuits.POSEIDON_M(t))
	p := bigIntMatrixToElements(circuits.POSEIDON_P(t))

	state := make([]fr.Element, t)
	state[0].SetBigInt(initialState)
	for i := 1; i < t; i++ {
		state[i].SetBigInt(inputs[i-1])
	}
	ark(&state, c, 0)

	for r := 0; r < nRoundsF/2-1; r++ {
		for j := 0; j < t; j++ {
			sigma(&state[j])
		}
		ark(&state, c, (r+1)*t)
		state = mix(state, m)
	}

	for j := 0; j < t; j++ {
		sigma(&state[j])
	}
	ark(&state, c, nRoundsF/2*t)
	state = mix(state, p)

	for r := 0; r < nRoundsP; r++ {
		sigma(&state[0])
		state[0].Add(&state[0], &c[(nRoundsF/2+1)*t+r])

		var newState0 fr.Element
		for j := 0; j < t; j++ {
			var mul fr.Element
			mul.Mul(&s[(t*2-1)*r+j], &state[j])
			newState0.Add(&newState0, &mul)
		}

		for k := 1; k < t; k++ {
			var mul fr.Element
			mul.Mul(&state[0], &s[(t*2-1)*r+t+k-1])
			state[k].Add(&state[k], &mul)
		}
		state[0] = newState0
	}

	for r := 0; r < nRoundsF/2-1; r++ {
		for j := 0; j < t; j++ {
			sigma(&state[j])
		}
		ark(&state, c, (nRoundsF/2+1)*t+nRoundsP+r*t)
		state = mix(state, m)
	}

	for j := 0; j < t; j++ {
		sigma(&state[j])
	}

	outs := make([]*big.Int, nOuts)
	for i := 0; i < nOuts; i++ {
		var out fr.Element
		mixLast(&out, state, m, i)
		outs[i] = elementToBigInt(out)
	}
	return outs
}

func sigma(x *fr.Element) {
	var x2 fr.Element
	x2.Square(x)
	var x4 fr.Element
	x4.Square(&x2)
	x.Mul(&x4, x)
}

func ark(state *[]fr.Element, c []fr.Element, r int) {
	s := *state
	for i := range s {
		s[i].Add(&s[i], &c[i+r])
	}
}

func mix(in []fr.Element, m [][]fr.Element) []fr.Element {
	out := make([]fr.Element, len(in))
	for col := range in {
		var acc fr.Element
		for row := range in {
			var term fr.Element
			term.Mul(&m[row][col], &in[row])
			acc.Add(&acc, &term)
		}
		out[col] = acc
	}
	return out
}

func mixLast(out *fr.Element, in []fr.Element, m [][]fr.Element, idx int) {
	out.SetZero()
	for row := range in {
		var term fr.Element
		term.Mul(&m[row][idx], &in[row])
		out.Add(out, &term)
	}
}

func bigIntSliceToElements(in []*big.Int) []fr.Element {
	out := make([]fr.Element, len(in))
	for i := range in {
		out[i].SetBigInt(in[i])
	}
	return out
}

func bigIntMatrixToElements(in [][]*big.Int) [][]fr.Element {
	out := make([][]fr.Element, len(in))
	for i := range in {
		out[i] = make([]fr.Element, len(in[i]))
		for j := range in[i] {
			out[i][j].SetBigInt(in[i][j])
		}
	}
	return out
}

func elementToBigInt(in fr.Element) *big.Int {
	var out big.Int
	in.BigInt(&out)
	return &out
}
