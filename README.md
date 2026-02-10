# Noema

Verifiable dataset compliance using Gemini reasoning and zero-knowledge proofs.

Noema enables organizations to evaluate private datasets against governance policies (privacy, safety, regulatory constraints) and generate a publicly verifiable proof that the policy decision was computed correctly—without ever revealing the sensitive dataset itself.
📸 Demo Video

Watch the 3-minute walkthrough
✨ Features

    🔒 ZK Compliance Proofs — Verify policy adherence without exposing raw data.

    🧠 Gemini-Powered Reasoning — High-level AI evaluation of complex datasets and images.

    ⚙️ Gnark Integration — Uses the gnark ZK-SNARK library for circuit generation.

    🖥️ Go Backend — High-performance orchestration of AI and ZK pipelines.

    📊 Web Dashboard — Clean UI for dataset evaluation and proof verification.

💡 Inspiration

Modern compliance requires a trade-off: share your private data with auditors, or skip verification. Noema flips this narrative. By combining LLM reasoning (to understand data) with Zero-Knowledge Proofs (to prove the result), we create a "Trust, but Verify" layer for private data.
🚀 How It Works

    Evaluation: Gemini analyzes the dataset (or images) against specific governance constraints.

    Aggregation: Policy results (pass/fail/severity) are converted into structured constraints.

    ZK Proof: A gnark circuit generates a proof that the final policy decision was derived correctly from the evaluation outputs.

    Verification: The proof is verified publicly. The verifier knows the policy was followed, but never sees the source data.

🛠 Tech Stack
Layer	Description
AI Engine	Gemini Pro / Flash for dataset reasoning
ZK Library	gnark (Groth16/PlonK) for circuit logic
Backend	Go (Golang) server
Frontend	React-based dashboard
📦 Setup & Run

Requirements

    Go 1.21+

    Gemini API Key

Quickstart
Bash

# Clone and enter backend
git clone https://github.com/your-repo/noema.git
cd noema/backend

# Configure environment
cp .env.example .env
# Edit .env and set JUDGE_KEY & GEMINI_API_KEY

# Run server
go run ./cmd/server

Server runs on: http://localhost:8080
🔒 API Summary

    POST /api/evaluate — (Gated) Processes dataset and generates ZK proof.

    POST /api/verify — (Public) Verifies a generated proof.

    GET /app — (Gated) Accesses the visual management dashboard.

🧠 What's Next

    ✅ On-chain Verification — Exporting proofs to EVM/Aleo smart contracts.

    📱 Edge Processing — Local evaluation to further minimize data transit.

    🌍 Multi-Model Consensus — Using multiple LLMs to reach a "verifiable consensus" on data safety.
