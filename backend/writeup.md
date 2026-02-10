## Inspiration

I’ve recently been deeply interested in AI safety, governance, and compliance. As more organizations train models on proprietary or sensitive datasets, a major challenge emerges: companies often need to demonstrate that their datasets comply with safety and regulatory standards, but they cannot reveal the datasets themselves due to privacy, intellectual property, or security concerns.

This led me to explore a simple idea: *What if organizations could privately evaluate their datasets using AI, and then generate a cryptographic proof that the compliance decision was computed correctly — without revealing the dataset or the detailed findings?*

With the availability of powerful reasoning models like Gemini and modern zero-knowledge proof systems, I saw an opportunity to combine **AI-assisted dataset evaluation** with **verifiable computation**, enabling trust without disclosure.

---

## What it does

**Noema** is a proof-of-concept system for privacy-preserving dataset compliance verification.

A dataset owner uploads a dataset and selects governance policies such as privacy, safety, or regulatory constraints. Gemini evaluates the dataset and produces structured severity assessments for each constraint. These assessments are then used as inputs to a zero-knowledge circuit that verifies that the policy decision (PASS / FAIL) was computed correctly from the declared policy configuration and evaluation outputs.

The resulting proof allows anyone to verify that the compliance result was derived honestly, while the dataset itself — and optionally the individual constraint results — remain private.

---

## How I built it

The system is implemented as a Golang-based backend platform integrating:

- Gemini API for structured dataset reasoning
- A policy evaluation pipeline that converts reasoning outputs into deterministic policy inputs
- gnark for constructing a Groth16 zero-knowledge policy gate circuit
- Cryptographic commitments linking dataset digests, policy configuration, and evaluation outputs
- A lightweight web interface for dataset upload, constraint configuration, evaluation runs, and proof verification

The backend processes datasets, runs AI evaluation, computes policy results, constructs a witness, generates a zero-knowledge proof, verifies the proof server-side, and exposes verification endpoints for external checking.

**Prototype note:**
Some dashboard elements and demonstration runs use controlled or simulated evaluation inputs to ensure stable demonstrations while the end-to-end automation pipeline is being refined. The zero-knowledge proof generation and verification pipeline, policy computation logic, and backend evaluation architecture are fully implemented.

---

## Challenges I ran into

Translating policy evaluation logic into circuit constraints required careful design to ensure deterministic witness construction and correct policy evaluation inside the proof system. Working with zero-knowledge circuits in a full evaluation pipeline also required handling key generation, constraint modeling, and integration challenges between AI evaluation outputs and cryptographic verification.

In parallel, I experimented extensively with new agent-assisted development workflows, learning how to effectively guide multi-agent coding processes while maintaining reliable and reproducible system behavior.

---

## Accomplishments that I'm proud of

- Building an end-to-end prototype combining **AI-assisted dataset evaluation** with **zero-knowledge verifiable policy computation**
- Implementing a working policy gate circuit capable of proving compliance decisions without revealing the underlying dataset
- Successfully bringing a long-standing research-style idea into a functional prototype within a hackathon timeframe

---

## What I learned

This project highlighted how **AI reasoning systems and cryptographic proof systems can complement each other**: AI provides semantic interpretation of complex datasets, while zero-knowledge proofs provide verifiable guarantees about how policy decisions are computed.

I also gained experience experimenting with emerging AI-assisted development workflows, which significantly accelerated the ability to prototype complex infrastructure concepts.

---

## What's next for Noema

Future work includes:

- Expanding multimodal dataset support (text, structured data, and images)
- Refining circuit ergonomics and improving proving performance
- Developing stronger standardized governance policy templates
- Exploring verifiable attestation systems for reasoning outputs and evaluation pipelines

Long-term, Noema explores a future where **AI dataset governance can be verified cryptographically while preserving privacy**, reducing reliance on trust-based compliance claims.
