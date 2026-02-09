package policyzk

import (
	"github.com/AlpinYukseloglu/poseidon-gnark/circuits"
	"github.com/consensys/gnark/frontend"
)

const N = 6

// PolicyGateCircuit proves:
//   - The prover knows (datasetDigest, policy config, gemini severities) that hash to Commitment
//   - And the deterministic policy check passes:
//     for each i: if Enabled[i] == 1 then Severity[i] <= MaxAllowed[i]
//
// Public outputs:
// - Commitment: binds everything (dataset + policy + gemini outputs)
// - OverallPass: 1 iff policy passes
// - MaxSeverity (optional): max observed severity among enabled constraints
type PolicyGateCircuit struct {
	// ===== Private witness =====
	// Dataset digest split into two limbs (suggest 128-bit each) to fit cleanly in Fr.
	// Backend must encode deterministically.
	DatasetDigestLo frontend.Variable
	DatasetDigestHi frontend.Variable

	// Policy config
	Enabled    [N]frontend.Variable
	MaxAllowed [N]frontend.Variable

	// Gemini outputs
	Severity [N]frontend.Variable

	// ===== Public signals =====
	Commitment  frontend.Variable `gnark:",public"` // Poseidon(domain, datasetDigest, enabled, maxAllowed, severity)
	OverallPass frontend.Variable `gnark:",public"` // boolean

	// Optional public signal (you can omit if you don’t want to reveal it)
	MaxSeverity frontend.Variable `gnark:",public"` // 0..2
}

func (c *PolicyGateCircuit) Define(api frontend.API) error {
	// --- constrain public outputs ---
	assertBoolean(api, c.OverallPass)
	assertIn012(api, c.MaxSeverity)

	// --- per-constraint checks + track max severity among enabled constraints ---
	anyFail := frontend.Variable(0)

	anySev2 := frontend.Variable(0) // OR over enabled * (severity==2)
	anySev1 := frontend.Variable(0) // OR over enabled * (severity==1)

	for i := 0; i < N; i++ {
		// ranges
		assertBoolean(api, c.Enabled[i])
		assertIn012(api, c.MaxAllowed[i])
		assertIn012(api, c.Severity[i])

		// indicators (exact, boolean, and partition-of-unity)
		s0, s1, s2 := indicators012(api, c.Severity[i])
		m0, m1, m2 := indicators012(api, c.MaxAllowed[i])
		_ = s0
		_ = m2

		// --- compute gt = (Severity > MaxAllowed) for enums 0..2 ---
		// Only possible gt cases:
		//  (1>0), (2>0), (2>1)
		gt := api.Add(
			api.Mul(s1, m0),
			api.Mul(s2, m0),
			api.Mul(s2, m1),
		)
		// gt is boolean automatically because these cases are mutually exclusive,
		// but we can still enforce boolean if you want:
		assertBoolean(api, gt)

		// fail_i = enabled_i * gt
		fail := api.Mul(c.Enabled[i], gt)
		anyFail = orBool(api, anyFail, fail)

		// --- track max severity among enabled constraints ---
		sev2Enabled := api.Mul(c.Enabled[i], s2)
		sev1Enabled := api.Mul(c.Enabled[i], s1)

		anySev2 = orBool(api, anySev2, sev2Enabled)
		anySev1 = orBool(api, anySev1, sev1Enabled)
	}

	// OverallPass = 1 - anyFail
	api.AssertIsEqual(c.OverallPass, api.Sub(1, anyFail))

	// MaxSeverity computed from (anySev2, anySev1):
	// if anySev2 => 2
	// else if anySev1 => 1
	// else 0
	// max = 2*anySev2 + (1-anySev2)*anySev1
	maxComputed := api.Add(
		api.Mul(2, anySev2),
		api.Mul(api.Sub(1, anySev2), anySev1),
	)
	api.AssertIsEqual(c.MaxSeverity, maxComputed)

	// --- commitment binding ---
	// Commitment = Poseidon(domainSep, datasetDigestLo, datasetDigestHi,
	//                       enabled[0..N-1], maxAllowed[0..N-1], severity[0..N-1])
	//
	// Domain separation prevents cross-protocol collisions.
	// Choose a fixed constant per circuit/version.
	const domainSep = 20260208 // pick ANY fixed int; version it if you change ordering/inputs

	inputs := make([]frontend.Variable, 0, 3+3*N)
	inputs = append(inputs, domainSep, c.DatasetDigestLo, c.DatasetDigestHi)
	for i := 0; i < N; i++ {
		inputs = append(inputs, c.Enabled[i])
	}
	for i := 0; i < N; i++ {
		inputs = append(inputs, c.MaxAllowed[i])
	}
	for i := 0; i < N; i++ {
		inputs = append(inputs, c.Severity[i])
	}

	commit := poseidonHashChunks(api, inputs)
	api.AssertIsEqual(c.Commitment, commit)

	return nil
}

// ===== Helpers =====

func assertBoolean(api frontend.API, x frontend.Variable) {
	// x*(x-1)=0 => x in {0,1}
	api.AssertIsEqual(api.Mul(x, api.Sub(x, 1)), 0)
}

func assertIn012(api frontend.API, x frontend.Variable) {
	// (x)(x-1)(x-2)=0 => x in {0,1,2}
	api.AssertIsEqual(
		api.Mul(x, api.Mul(api.Sub(x, 1), api.Sub(x, 2))),
		0,
	)
}

// indicators012 returns (eq0, eq1, eq2) for x constrained to {0,1,2}.
// Uses Lagrange basis polynomials (exact 0/1 in the field).
func indicators012(api frontend.API, x frontend.Variable) (frontend.Variable, frontend.Variable, frontend.Variable) {
	// inv2 = 1/2 in field
	inv2 := api.Inverse(2)

	// eq0 = (x-1)(x-2) / 2
	eq0 := api.Mul(inv2, api.Mul(api.Sub(x, 1), api.Sub(x, 2)))

	// eq1 = -x(x-2)
	eq1 := api.Neg(api.Mul(x, api.Sub(x, 2)))

	// eq2 = x(x-1) / 2
	eq2 := api.Mul(inv2, api.Mul(x, api.Sub(x, 1)))

	// They should be boolean and sum to 1 (good “production” invariant)
	assertBoolean(api, eq0)
	assertBoolean(api, eq1)
	assertBoolean(api, eq2)
	api.AssertIsEqual(api.Add(eq0, eq1, eq2), 1)

	return eq0, eq1, eq2
}

func poseidonHashChunks(api frontend.API, inputs []frontend.Variable) frontend.Variable {
	const maxInputs = 16
	if len(inputs) <= maxInputs {
		return circuits.Poseidon(api, inputs)
	}
	h := circuits.Poseidon(api, inputs[:maxInputs])
	rest := make([]frontend.Variable, 0, 1+len(inputs[maxInputs:]))
	rest = append(rest, h)
	rest = append(rest, inputs[maxInputs:]...)
	return circuits.Poseidon(api, rest)
}

// OR for booleans: a OR b = a + b - ab
func orBool(api frontend.API, a, b frontend.Variable) frontend.Variable {
	// (a,b) must already be boolean. We rely on that invariant.
	return api.Sub(api.Add(a, b), api.Mul(a, b))
}
