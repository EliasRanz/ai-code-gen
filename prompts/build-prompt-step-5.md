# Step 5: AI Generation Service and LLM Abstraction Layer

## Instructions

Implement the AI Generation Service and LLM Abstraction Layer as described. Provide the SSE endpoint, interface definitions, and stub out the VLLM client. Do not implement actual LLM calls yet.

- AI Generation Service:
  - Implement the `/generate/stream` SSE endpoint (authentication required).
  - Use a Go channel to stream stubbed responses.
  - Integrate Redis Pub/Sub for horizontal scaling (stub).
- LLM Abstraction Layer:
  - Define the `LLMClient` interface and related structs.
  - Stub out the VLLM client implementation.
- Do not implement actual LLM calls yet.
