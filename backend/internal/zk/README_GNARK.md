# ZK Stub Flow + Gnark (BN254) Sketch

## How the current stub works

### Call flow
1. POST /api/evaluate -> backend/internal/evaluate/handler.go
2. Build commitment = sha256(spec_json + eval_json) and compute policy outputs.
3. Call zk.GenerateProof(PublicInputs{...}).
4. Return proof + public_inputs in API response.
5. POST /api/verify -> backend/internal/verify/handler.go
6. Call zk.VerifyProof(proof_b64, public_inputs_b64).

### What GenerateProof receives
GenerateProof gets a PublicInputs struct:

- PolicyThreshold: int (0..2)
- MaxSeverity: int (0..2)
- OverallPass: bool
- Commitment: string (0x-prefixed hex)

These are derived from:
- policyThreshold, maxSeverity, overallPass from computePolicyResult(...)
- commitment from zk.CommitmentSHA256([]byte("spec"), specJSON, []byte("eval"), evalJSON)

### What VerifyProof receives
VerifyProof gets two base64 strings:
- proof_b64
- public_inputs_b64

These are the values previously returned by GenerateProof, or submitted later by a client.

### Public inputs encoding (stub format)
The PublicInputs struct is encoded as UTF-8 bytes:

noema_public_inputs_v1|pt=<int>|ms=<int>|op=<0|1>|c=<hex>

public_inputs_b64 = base64(UTF-8 bytes above)

### Stub proof format
The stub proof is a SHA-256 checksum of the public inputs:

proof_raw = "noema_stub_proof_v1|" + sha256_hex(public_inputs_bytes)
proof_b64 = base64(proof_raw)

### Stub verification
Verification recomputes the checksum from public_inputs and checks it matches proof_raw.

## Concrete example

Given:
- PolicyThreshold = 1
- MaxSeverity = 0
- OverallPass = true
- Commitment = 0xdeadbeef

Public inputs string:
noema_public_inputs_v1|pt=1|ms=0|op=1|c=0xdeadbeef

public_inputs_b64 = base64(above)
proof_raw = "noema_stub_proof_v1|" + sha256_hex(public_inputs_bytes)
proof_b64 = base64(proof_raw)

## Gnark (BN254) integration sketch

### Design goals
- Keep the public API shape unchanged: proof_b64 + public_inputs_b64.
- Use gnark to generate and verify a real proof using BN254.
- Map the existing public inputs into circuit public inputs in the same order.

### Proposed wire format
Option A (recommended): pack public inputs as 4 field elements in this order:
1. policy_threshold (0..2)
2. max_severity (0..2)
3. overall_pass (0 or 1)
4. commitment (big integer parsed from 0x hex)

Encode public_inputs_b64 as a compact JSON array or binary (your choice), but keep
DecodePublicInputs returning the same logical values. The client does not need to
understand the gnark field encoding; it can keep submitting the base64 blob.

Option B: keep the current UTF-8 string format for public_inputs_b64, and inside
GenerateProof parse it into field elements for gnark. This is simplest to keep
backwards compatibility for the public verify endpoint.

### Circuit sketch (gnark)

- Inputs:
  - Public: pt, ms, op, commitment
  - Private: whatever witness proves that commitment matches the private data

- Constraints:
  - Enforce pt in [0..2], ms in [0..2], op in {0,1}
  - Enforce commitment is hash(spec_json, eval_json) inside the circuit, or
    equivalently enforce a relationship that binds private inputs to commitment.

### Go API sketch

Keep the existing signatures in zk.go:

func GenerateProof(pi PublicInputs) (Proof, error)
func VerifyProof(proofB64, publicInputsB64 string) (bool, string, error)

Implement gnark under the hood:

- GenerateProof:
  - Parse pi.Commitment (0x hex) into big.Int
  - Build public inputs in order: pt, ms, op, commitment
  - Build witness with private inputs (spec + eval data, or hashes derived from them)
  - Run groth16.Prove and return proof bytes (base64)
  - Encode public inputs into public_inputs_b64 (string or binary)

- VerifyProof:
  - Decode proof bytes
  - Decode public inputs bytes into field elements in the same order
  - Run groth16.Verify

### Example skeleton (pseudocode)

func GenerateProof(pi PublicInputs) (Proof, error) {
  // 1) Build public inputs
  pub := []fr.Element{
    fr.NewElement(uint64(pi.PolicyThreshold)),
    fr.NewElement(uint64(pi.MaxSeverity)),
    fr.NewElement(uint64(boolToInt(pi.OverallPass))),
    fr.ElementFromBigInt(parseCommitment(pi.Commitment)),
  }

  // 2) Build witness with private inputs
  witness := MyCircuitWitness{ /* spec/eval data */ }
  full, err := frontend.NewWitness(&witness, ecc.BN254)

  // 3) Prove
  proof, err := groth16.Prove(r1cs, pk, full)

  // 4) Serialize proof and public inputs
  proofB64 := base64.StdEncoding.EncodeToString(serializeProof(proof))
  publicInputsB64 := base64.StdEncoding.EncodeToString(serializePublicInputs(pub))

  return Proof{System: "groth16", Curve: "bn254", ProofB64: proofB64, PublicInputsB64: publicInputsB64}, nil
}

func VerifyProof(proofB64, publicInputsB64 string) (bool, string, error) {
  proof := deserializeProof(base64Decode(proofB64))
  pub := deserializePublicInputs(base64Decode(publicInputsB64))
  ok, err := groth16.Verify(proof, vk, pub)
  if err != nil { return false, "verify error", err }
  if !ok { return false, "invalid proof", nil }
  return true, "verified", nil
}

### Where to wire gnark
- Replace the internals of `backend/internal/zk/zk.go`.
- Keep the existing `PublicInputs` struct and Encode/Decode helpers if you want
  compatibility. Or swap them for a gnark-optimized public inputs format.

### Files to edit
- backend/internal/zk/zk.go
- backend/internal/zk/FORMAT.md (if you change public_inputs_b64 format)
- backend/internal/evaluate/handler.go (only if you change how PublicInputs are built)
- backend/internal/verify/handler.go (only if you change API structure)

