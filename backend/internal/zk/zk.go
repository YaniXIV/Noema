package zk

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const (
	ProofSystem = "groth16"
	ProofCurve  = "bn254"

	publicInputsPrefix = "noema_public_inputs_v1|"
	proofPrefix        = "noema_stub_proof_v1|"
)

// PublicInputs define the public inputs for policy aggregation.
// Format (UTF-8 bytes):
// noema_public_inputs_v1|pt=<int>|ms=<int>|op=<0|1>|c=<hex commitment>
//
// Commitment is a hex string with 0x prefix.
// Thresholds and severities are 0..2.
// Overall pass is 0 or 1.
//
// This format is stable and suitable for the stub proof pipeline.
// The full circuit should produce the same public inputs ordering.
type PublicInputs struct {
	PolicyThreshold int
	MaxSeverity     int
	OverallPass     bool
	Commitment      string
}

// Proof bundles the base64-encoded proof and public inputs.
type Proof struct {
	System          string
	Curve           string
	ProofB64        string
	PublicInputsB64 string
}

func GenerateProof(pi PublicInputs) (Proof, error) {
	pub, err := EncodePublicInputs(pi)
	if err != nil {
		return Proof{}, err
	}
	proofRaw := proofPrefix + hexDigest(pub)
	return Proof{
		System:          ProofSystem,
		Curve:           ProofCurve,
		ProofB64:        base64.StdEncoding.EncodeToString([]byte(proofRaw)),
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
	pub, err := base64.StdEncoding.DecodeString(publicInputsB64)
	if err != nil {
		return false, "invalid public inputs encoding", fmt.Errorf("invalid public inputs encoding")
	}
	if !strings.HasPrefix(string(pub), publicInputsPrefix) {
		return false, "invalid public inputs format", nil
	}
	if !strings.HasPrefix(string(proofRaw), proofPrefix) {
		return false, "invalid proof format", nil
	}
	expected := proofPrefix + hexDigest(pub)
	if expected != string(proofRaw) {
		return false, "proof does not match public inputs", nil
	}
	return true, "stub verifier", nil
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

func hexDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
