# Noema

**Verifiable dataset compliance using Gemini reasoning and zero-knowledge proofs.**

Noema is a proof-of-concept system that enables organizations to evaluate private datasets against governance policies (privacy, safety, regulatory constraints) and generate a **publicly verifiable proof that the policy decision was computed correctly**, without revealing the underlying dataset.

📸 Demo Video  
https://www.youtube.com/watch?v=VvV0VkkP7tY

---

## Features

- 🔒 **ZK compliance proofs** — verify policy decision correctness without exposing raw data  
- 🧠 **Gemini-assisted dataset evaluation** — structured reasoning over datasets (and optional images)  
- ⚙️ **gnark integration** — Groth16 policy gate circuit for proof generation  
- 🖥️ **Go backend** — orchestration of evaluation, policy computation, and proof generation  
- 📊 **Web dashboard** — interface for dataset upload, constraint configuration, and proof verification  

---

## Inspiration

Modern compliance workflows often require organizations to reveal sensitive datasets to auditors in order to demonstrate adherence to safety or regulatory standards. This creates a fundamental trade-off between **verification** and **privacy**.

Noema explores an alternative approach: using AI systems to evaluate datasets and zero-knowledge proofs to verify that the resulting policy decision was computed honestly, enabling **trust without disclosure**.

---

## How it works

1. A dataset owner uploads a dataset and selects governance constraints.
2. Gemini evaluates the dataset and produces structured severity assessments for each constraint.
3. Policy aggregation computes the final policy decision (PASS / FAIL).
4. A gnark zero-knowledge circuit generates a proof verifying that the policy decision was computed correctly from the declared inputs.
5. Anyone can verify the proof without accessing the underlying dataset.

**Important:**  
The system proves that the policy decision was computed correctly from the declared evaluation outputs. It does not prove that the AI evaluation itself is correct.

---

## Tech stack

| Layer | Technology |
|------|-----------|
| AI evaluation | Gemini API |
| ZK proofs | gnark (Groth16) |
| Backend | Go (Gin) |
| Frontend | Server-rendered web interface |

---

## Setup & Run

Requirements:
- Go 1.21+
- (Optional) Gemini API key

```bash
git clone https://github.com/your-repo/noema.git
cd noema/backend

cp .env.example .env
# Set JUDGE_KEY and GEMINI_API_KEY

go run ./cmd/server
