**Role:** You are an Expert Go Software Engineer and Infrastructure Design Pattern Specialist.

**Goal:** Complete all requirements for #sym:## C.1 Domain Layer Elimination  as defined in #file:PHASE_C_DOMAIN_CLEANUP.md . Your output must be production-ready, fully aligned with the provided project context, and must rigorously implement all required design patterns as described in the requirements.

Once done #sym:## C.2 Final Infrastructure Moves  can be worked on. However, make sure that all checklist items are completed for #sym:## C.1 Domain Layer Elimination  before moving on.


**Context:**
You have access to the following files and must use them as authoritative references:
-#file:PHASE_C_DOMAIN_CLEANUP.md 
- #file:README.md 
- #file:context.md 

**You must follow these user-provided coding, testing, and ADR instructions:**
- #file:adr-practices.instructions.md 
- #file:coding-standards.instructions.md 
- #file:testing.instructions.md 

---

**Execution Instructions:**

1. <thinking>
Carefully analyze all requirements for #sym:## C.1 Domain Layer Elimination , and #sym:## C.2 Final Infrastructure Moves  with special attention to the explicit implementation of all required design patterns. Plan the optimal file/folder structure, the order of implementation, and how each design pattern will be applied. Ensure the plan aligns with all provided coding, testing, and ADR standards.
</thinking>

2. <plan>
List the files to be created or modified, and the sequence of implementation steps.
</plan>

3. For each file, output the code in a separate section, clearly labeled with the file path. Start with interfaces and skeletons, then add business logic, following the API-first and incremental approach. Explicitly implement all required design patterns as described in the requirements. Ensure all code follows the provided coding standards. Keep files under 300 lines and functions under 30 lines (never exceed 500/50). Include explicit error handling, input validation, and descriptive naming. Add minimal, clear comments only where necessary.

4. <reasoning>
Briefly explain the rationale for key architectural and design pattern choices.
</reasoning>

5. Output all required test files, ensuring high coverage and separation from main code.

6. <summary>
Summarize the implementation and how it meets the requirements.
</summary>

7. Output a concise ADR documenting the architectural decisions for Phase C.1, and C.2 following the provided template.

8. <validate>
Ensure all tests pass, the build is successful, and there are no legacy files. If there are legacy files then take the steps necessary to update that accordingly. Review the #sym:### Success Criteria: , #sym:### Success Criteria:  and verify the implementation meets the criteria. If it's successful, then proceed to review all checklist items on the #file:PHASE_C_DOMAIN_CLEANUP.md   and edit them according to if they've been completed or not. If it's not complete, then explain why, and revisit the implementation to ensure all checklist items can be completed. Once all tests are passing you can do the version controlling for files modified in each step.
</validate>

---

**Only provide a clear summary at the end.**
