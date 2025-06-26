# AI System Generation Prompt: Full-Stack AI UI Generator

## Agent Prompts for Iterative System Generation

Use the following prompts for each step. For each, include the referenced file using `#file:build-prompt-step-*.md` (replace * with the step number or letter as appropriate):

---

**Step 0: Project Overview and Usage Instructions**
> Please read and understand the project overview and instructions in #file:build-prompt-step-0.md. Use this as context for all subsequent steps.

**Step 1: Project Structure, Tech Stack, and High-Level Goal**
> Generate the directory and file structure for the monorepo as described in #file:build-prompt-step-1.md. Provide placeholder content for all files. Do not implement any business logic yet.

**Step 2: Backend Service Scaffolding and Shared Libraries**
> Implement the entrypoint files for each Go service, the shared observability package, and configuration loading as described in #file:build-prompt-step-2.md. Ensure each service can start, log, and shut down gracefully. Do not implement any business logic or API endpoints yet.

**Step 3: API Gateway, Auth, and User/Project Service Stubs**
> Implement the API Gateway, Authentication Service, and User/Project Service as described in #file:build-prompt-step-3.md. Define all routes, middleware, and gRPC interfaces, but do not implement business logic or database access yet.

**Step 4a: Database Schema and Migrations**
> Write the SQL schema and migration scripts for users, projects, chat_sessions, and chat_messages as described in #file:build-prompt-step-4a.md. Do not implement repository code or CRUD logic yet.

**Step 4b: Repository Pattern and CRUD Stubs**
> Implement the repository pattern and CRUD stubs for users and projects as described in #file:build-prompt-step-4b.md. Do not implement full business logic or validation yet.

**Step 5: AI Generation Service and LLM Abstraction Layer**
> Implement the AI Generation Service and LLM Abstraction Layer as described in #file:build-prompt-step-5.md. Provide the SSE endpoint, interface definitions, and stub out the VLLM client. Do not implement actual LLM calls yet.

**Step 6a: Next.js App Structure and Routing**
> Scaffold the Next.js app structure and routing as described in #file:build-prompt-step-6a.md. Do not implement UI components or business logic yet.

**Step 6b: ChatInterface, PreviewPane, and SSE Client**
> Implement placeholder components and a basic SSE client as described in #file:build-prompt-step-6b.md. Do not implement detailed UI or business logic yet.

**Step 6c: Authentication Flow Stubs**
> Scaffold the authentication flow for the frontend as described in #file:build-prompt-step-6c.md. Do not implement full authentication logic yet.

**Step 7: DevOps - Docker, Compose, and CI/CD**
> Generate Dockerfiles for each service, a docker-compose.yml, and a GitHub Actions CI/CD workflow as described in #file:build-prompt-step-7.md.

**Step 8: Incremental Business Logic, Integration, and Testing**
> For each service, incrementally implement business logic, integration, and tests as described in #file:build-prompt-step-8.md. Focus on one feature or service at a time, and ensure all tests pass before proceeding.

**Step 9: Final Integration and End-to-End (E2E) Tests**
> Integrate all services and features, and implement end-to-end tests as described in #file:build-prompt-step-9.md. Ensure the system works as a whole and all tests pass.

---

## Success Rate Analysis for This Iterative Build Process

**Analysis provided by: GitHub Copilot (OpenAI/Gemini/Claude-class LLM agent, June 2025)**

### Expected Success Rate

- **Scaffolding, structure, and basic service setup:** ~100%
- **API/gRPC stubs, database schema, and DevOps:** ~95–100%
- **Business logic, integration, and E2E:** ~80–90% (may require minor manual fixes or clarifications)
- **Fully working, production-ready system in one pass:** ~80–90% (with some manual iteration likely needed for edge cases, integration bugs, or advanced features)

### Why This Works
- Each step is focused and within LLM context limits, reducing confusion and hallucination.
- Validation and testing are built into the process, catching errors early.
- The process allows for retrying or splitting any step that fails, minimizing blockers.
- Explicit instructions to use placeholders/stubs prevent premature complexity.
- Reference to a detailed architecture ensures alignment with your goals.

### Remaining Limitations
- Any remaining gap to 100% is due to current LLM limitations:
  - Context window size (can’t “see” the whole system at once)
  - Occasional hallucination or misunderstanding of requirements
  - Difficulty with complex cross-service integration or edge cases
  - Inconsistent handling of advanced error handling, security, or non-standard patterns
  - Occasional syntax or logic errors in generated code

### Source
This analysis is provided by GitHub Copilot (OpenAI/Gemini/Claude-class LLM agent, June 2025), based on current LLM capabilities and best practices for iterative, validated code generation.
