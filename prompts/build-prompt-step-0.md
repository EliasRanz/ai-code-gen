# Step 0: Project Overview and Usage Instructions

## Project Overview

This project is a production-ready, full-stack AI UI Generation System inspired by Vercel's v0.dev. It is designed to transform natural language prompts into high-quality, interactive frontend components using a modular, scalable, and observable microservices architecture. The backend is written in Go (Gin), the frontend in Next.js (TypeScript, Tailwind CSS, shadcn/ui), and the system includes robust authentication, testing, and DevOps practices.

## How to Use the Stepwise Prompts

- The system is built iteratively using a series of stepwise prompts (`build-prompt-step-*.md`).
- Each step focuses on a specific aspect of the system, from scaffolding to business logic and integration.
- For each step, use the `#fetch:file` feature to include the relevant markdown file in your Copilot/LLM chat, and follow the instructions provided.
- **After each step, validate the output** (e.g., does the code compile, do tests pass, does the structure match expectations?) before proceeding to the next.
- **Explicitly remind the agent in each step:**
  - Use only scaffolding, stubs, and placeholders unless otherwise specified.
  - Do not implement business logic or detailed UI until the appropriate step.
- If you encounter issues, consider splitting a step into smaller sub-steps (e.g., frontend routing vs. components). For example, split database schema and repository into 4a/4b, or frontend into 6a/6b/6c.
- **After each major step, prompt the agent to generate a checklist or minimal test to verify the output** (e.g., "Write a test to ensure the service starts and logs output").
- **Final Integration:** After all features are implemented, use a final step for integration, wiring, and end-to-end testing to ensure the system works as a whole.

## Reference Architecture

For all steps, refer to the detailed architecture and rationale in `gemini-report.md` for context, design decisions, and requirements.

- If you need clarification or details about any part of the system, consult `gemini-report.md`.
- The LLM prompt for code generation is in `build-prompt.md`.
- When copying content from the report, use standard Markdown headings (no bold) and remove citation numbers unless needed for internal reference.

## Step List (with possible sub-steps)

1. `build-prompt-step-0.md` — Project overview and usage instructions (this file)
2. `build-prompt-step-1.md` — Project structure, tech stack, high-level goal
3. `build-prompt-step-2.md` — Backend service scaffolding and shared libraries
4. `build-prompt-step-3.md` — API Gateway, Auth, and User/Project service stubs
5. `build-prompt-step-4a.md` — Database schema and migrations
6. `build-prompt-step-4b.md` — Repository pattern and CRUD stubs
7. `build-prompt-step-5.md` — AI Generation Service and LLM abstraction layer
8. `build-prompt-step-6a.md` — Next.js app structure and routing
9. `build-prompt-step-6b.md` — ChatInterface, PreviewPane, SSE client
10. `build-prompt-step-6c.md` — Authentication flow stubs
11. `build-prompt-step-7.md` — DevOps: Docker, Compose, and CI/CD
12. `build-prompt-step-8.md` — Incremental business logic, integration, and testing
13. `build-prompt-step-9.md` — Final integration and end-to-end (E2E) tests

---

**Start with Step 1 and proceed sequentially. Use this file and `gemini-report.md` as your reference throughout the process.**

---

## Additional Best Practices for Success

- **Validation:** After each step, generate a checklist or minimal test to verify the output (e.g., "Write a test to ensure the service starts and logs output").
- **Explicit Placeholders:** Instruct the agent to use placeholders and not implement business logic or detailed UI too early.
- **Sub-Steps:** If a step is too large or fails, break it into smaller, focused sub-steps (e.g., 4a/4b, 6a/6b/6c).
- **Final Integration:** After all features are implemented, use a final step for integration, wiring, and end-to-end testing to ensure the system works as a whole.
- **Markdown Consistency:** Use standard Markdown headings and formatting for all generated files.
