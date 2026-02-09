package zk

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/AlpinYukseloglu/poseidon-gnark/circuits"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"

	"noema/internal/zk/policyzk"
)

const (
	ProofSystem = "groth16"
	ProofCurve  = "bn254"

	publicInputsPrefix = "noema_public_inputs_v1|"
	policyGateDomainSep = 20260208
)

// PolicyGateConstraintCount is the fixed N for the PolicyGateCircuit.
const PolicyGateConstraintCount = policyzk.N

// PublicInputs define the public inputs for policy aggregation.
// Format (UTF-8 bytes):
// noema_public_inputs_v1|pt=<int>|ms=<int>|op=<0|1>|c=<hex commitment>
//
// Commitment is a hex string with 0x prefix.
// Thresholds and severities are 0..2.
// Overall pass is 0 or 1.
//
// This format is stable and suitable for the public API.
type PublicInputs struct {
	PolicyThreshold int
	MaxSeverity     int
	OverallPass     bool
	Commitment      string

	// Witness is required for proof generation.
	Witness *WitnessInputs
}

// WitnessInputs carries the private inputs for the PolicyGateCircuit.
type WitnessInputs struct {
	DatasetDigestHex string
	Enabled          [PolicyGateConstraintCount]uint64
	MaxAllowed       [PolicyGateConstraintCount]uint64
	Severity         [PolicyGateConstraintCount]uint64
}

// Proof bundles the base64-encoded proof and public inputs.
type Proof struct {
	System          string
	Curve           string
	ProofB64        string
	PublicInputsB64 string
}

var (
	initOnce  sync.Once
	initErr   error
	cachedR1CS constraint.ConstraintSystem
	cachedPK  groth16.ProvingKey
	cachedVK  groth16.VerifyingKey
)

func initGroth16() error {
	initOnce.Do(func() {
		var circuit policyzk.PolicyGateCircuit
		cachedR1CS, initErr = frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if initErr != nil {
			return
		}
		cachedPK, cachedVK, initErr = groth16.Setup(cachedR1CS)
	})
	return initErr
}

func GenerateProof(pi PublicInputs) (Proof, error) {
	pub, err := EncodePublicInputs(pi)
	if err != nil {
		return Proof{}, err
	}
	if pi.Witness == nil {
		return Proof{}, fmt.Errorf("missing witness inputs")
	}
	if err := initGroth16(); err != nil {
		return Proof{}, err
	}

	commitmentInt, err := parseCommitmentHex(pi.Commitment)
	if err != nil {
		return Proof{}, err
	}

	computedCommitment, err := CommitmentPoseidon(pi.Witness.DatasetDigestHex, pi.Witness.Enabled, pi.Witness.MaxAllowed, pi.Witness.Severity)
	if err != nil {
		return Proof{}, err
	}
	if !strings.EqualFold(computedCommitment, pi.Commitment) {
		return Proof{}, fmt.Errorf("commitment does not match witness inputs")
	}

	datasetLo, datasetHi, err := datasetDigestLimbs(pi.Witness.DatasetDigestHex)
	if err != nil {
		return Proof{}, err
	}

	assignment := policyzk.PolicyGateCircuit{
		DatasetDigestLo: datasetLo,
		DatasetDigestHi: datasetHi,
		Enabled:         toVarArray(pi.Witness.Enabled),
		MaxAllowed:      toVarArray(pi.Witness.MaxAllowed),
		Severity:        toVarArray(pi.Witness.Severity),
		Commitment:      commitmentInt,
		OverallPass:     boolToInt(pi.OverallPass),
		MaxSeverity:     pi.MaxSeverity,
	}

	fullWitness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return Proof{}, err
	}

	proof, err := groth16.Prove(cachedR1CS, cachedPK, fullWitness)
	if err != nil {
		return Proof{}, err
	}

	var proofBuf bytes.Buffer
	if _, err := proof.WriteTo(&proofBuf); err != nil {
		return Proof{}, err
	}

	return Proof{
		System:          ProofSystem,
		Curve:           ProofCurve,
		ProofB64:        base64.StdEncoding.EncodeToString(proofBuf.Bytes()),
		PublicInputsB64: base64.StdEncoding.EncodeToString(pub),
	}, nil
}

func VerifyProof(proofB64, publicInputsB64 string) (bool, string, error) {
	if proofB64 == "" || publicInputsB64 == "" {
		return false, "missing proof or public inputs", fmt.Errorf("missing proof or public inputs")
	}
	proofRaw, err := base64.StdEncoding.DecodeString(proofB64)
	if err != nil {
		return false, "invalid proof encoding", fmt.Errorf("invalid proof encoding")
	}
	pubRaw, err := base64.StdEncoding.DecodeString(publicInputsB64)
	if err != nil {
		return false, "invalid public inputs encoding", fmt.Errorf("invalid public inputs encoding")
	}
	pi, err := DecodePublicInputs(pubRaw)
	if err != nil {
		return false, "invalid public inputs format", nil
	}
	if err := initGroth16(); err != nil {
		return false, "verifier init failed", err
	}

	commitmentInt, err := parseCommitmentHex(pi.Commitment)
	if err != nil {
		return false, "invalid commitment", err
	}

	assignment := policyzk.PolicyGateCircuit{
		Commitment:  commitmentInt,
		OverallPass: boolToInt(pi.OverallPass),
		MaxSeverity: pi.MaxSeverity,
	}
	publicWitness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return false, "invalid public witness", err
	}

	proof := groth16.NewProof(ecc.BN254)
	if _, err := proof.ReadFrom(bytes.NewReader(proofRaw)); err != nil {
		return false, "invalid proof encoding", err
	}

	if err := groth16.Verify(proof, cachedVK, publicWitness); err != nil {
		return false, "invalid proof", nil
	}
	return true, "verified", nil
}

func EncodePublicInputs(pi PublicInputs) ([]byte, error) {
	if pi.PolicyThreshold < 0 || pi.PolicyThreshold > 2 {
		return nil, fmt.Errorf("policy threshold must be 0..2")
	}
	if pi.MaxSeverity < 0 || pi.MaxSeverity > 2 {
		return nil, fmt.Errorf("max severity must be 0..2")
	}
	if pi.Commitment == "" {
		return nil, fmt.Errorf("commitment required")
	}
	op := 0
	if pi.OverallPass {
		op = 1
	}
	payload := fmt.Sprintf("%spt=%d|ms=%d|op=%d|c=%s", publicInputsPrefix, pi.PolicyThreshold, pi.MaxSeverity, op, pi.Commitment)
	return []byte(payload), nil
}

func DecodePublicInputs(pub []byte) (PublicInputs, error) {
	s := string(pub)
	if !strings.HasPrefix(s, publicInputsPrefix) {
		return PublicInputs{}, fmt.Errorf("invalid public inputs prefix")
	}
	fields := strings.Split(strings.TrimPrefix(s, publicInputsPrefix), "|")
	out := PublicInputs{}
	seenPT := false
	seenMS := false
	seenOP := false
	seenC := false
	for _, f := range fields {
		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			return PublicInputs{}, fmt.Errorf("invalid public inputs field")
		}
		switch kv[0] {
		case "pt":
			if seenPT {
				return PublicInputs{}, fmt.Errorf("duplicate policy threshold")
			}
			v, err := strconv.Atoi(kv[1])
			if err != nil {
				return PublicInputs{}, fmt.Errorf("invalid policy threshold")
			}
			if v < 0 || v > 2 {
				return PublicInputs{}, fmt.Errorf("policy threshold must be 0..2")
			}
			out.PolicyThreshold = v
			seenPT = true
		case "ms":
			if seenMS {
				return PublicInputs{}, fmt.Errorf("duplicate max severity")
			}
			v, err := strconv.Atoi(kv[1])
			if err != nil {
				return PublicInputs{}, fmt.Errorf("invalid max severity")
			}
			if v < 0 || v > 2 {
				return PublicInputs{}, fmt.Errorf("max severity must be 0..2")
			}
			out.MaxSeverity = v
			seenMS = true
		case "op":
			if seenOP {
				return PublicInputs{}, fmt.Errorf("duplicate overall pass")
			}
			v, err := strconv.Atoi(kv[1])
			if err != nil {
				return PublicInputs{}, fmt.Errorf("invalid overall pass")
			}
			if v != 0 && v != 1 {
				return PublicInputs{}, fmt.Errorf("overall pass must be 0 or 1")
			}
			out.OverallPass = v == 1
			seenOP = true
		case "c":
			if seenC {
				return PublicInputs{}, fmt.Errorf("duplicate commitment")
			}
			if kv[1] == "" {
				return PublicInputs{}, fmt.Errorf("commitment required")
			}
			if !strings.HasPrefix(kv[1], "0x") {
				return PublicInputs{}, fmt.Errorf("commitment must have 0x prefix")
			}
			if _, err := hex.DecodeString(strings.TrimPrefix(kv[1], "0x")); err != nil {
				return PublicInputs{}, fmt.Errorf("commitment must be hex")
			}
			out.Commitment = kv[1]
			seenC = true
		default:
			return PublicInputs{}, fmt.Errorf("unknown public inputs field")
		}
	}
	if !seenPT || !seenMS || !seenOP || !seenC {
		return PublicInputs{}, fmt.Errorf("missing public inputs field")
	}
	return out, nil
}

func CommitmentSHA256(parts ...[]byte) string {
	h := sha256.New()
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		_, _ = h.Write(p)
	}
	return "0x" + hex.EncodeToString(h.Sum(nil))
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func parseCommitmentHex(commitment string) (*big.Int, error) {
	if !strings.HasPrefix(commitment, "0x") {
		return nil, fmt.Errorf("commitment must have 0x prefix")
	}
	hexStr := strings.TrimPrefix(commitment, "0x")
	if hexStr == "" {
		return nil, fmt.Errorf("commitment required")
	}
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("commitment must be hex")
	}
	return new(big.Int).SetBytes(b), nil
}

func datasetDigestLimbs(digestHex string) (*big.Int, *big.Int, error) {
	b, err := hex.DecodeString(digestHex)
	if err != nil {
		return nil, nil, fmt.Errorf("dataset digest must be hex")
	}
	if len(b) != 32 {
		return nil, nil, fmt.Errorf("dataset digest must be 32 bytes")
	}
	hi := new(big.Int).SetBytes(b[:16])
	lo := new(big.Int).SetBytes(b[16:])
	return lo, hi, nil
}

func toVarArray(vals [PolicyGateConstraintCount]uint64) [PolicyGateConstraintCount]frontend.Variable {
	var out [PolicyGateConstraintCount]frontend.Variable
	for i := 0; i < PolicyGateConstraintCount; i++ {
		out[i] = vals[i]
	}
	return out
}

// CommitmentPoseidon computes the PolicyGateCircuit commitment.
func CommitmentPoseidon(datasetDigestHex string, enabled, maxAllowed, severity [PolicyGateConstraintCount]uint64) (string, error) {
	lo, hi, err := datasetDigestLimbs(datasetDigestHex)
	if err != nil {
		return "", err
	}

	inputs := make([]*big.Int, 0, 3+3*PolicyGateConstraintCount)
	inputs = append(inputs, big.NewInt(policyGateDomainSep))
	inputs = append(inputs, lo, hi)
	for i := 0; i < PolicyGateConstraintCount; i++ {
		inputs = append(inputs, new(big.Int).SetUint64(enabled[i]))
	}
	for i := 0; i < PolicyGateConstraintCount; i++ {
		inputs = append(inputs, new(big.Int).SetUint64(maxAllowed[i]))
	}
	for i := 0; i < PolicyGateConstraintCount; i++ {
		inputs = append(inputs, new(big.Int).SetUint64(severity[i]))
	}

	hash := poseidonHashChunksNative(inputs)
	hexStr := hash.Text(16)
	if len(hexStr)%2 == 1 {
		hexStr = "0" + hexStr
	}
	return "0x" + hexStr, nil
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
