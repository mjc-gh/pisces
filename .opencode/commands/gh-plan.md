---
description: Refine a GitHub issue with implementation plan and checklist
model: openrouter/anthropic/claude-opus-4.5
---

Fetch GitHub issue #$1 and refine it with a detailed implementation plan.

**Instructions:**

1. Fetch the current issue content:
   `gh issue view $1 --repo mjc-gh/pisces`

2. Analyze the issue to understand:
   - What problem needs to be solved
   - What features or fixes are requested
   - Any constraints or requirements mentioned
   - Ask clarifying questions if needed

3. Explore the codebase thoroughly to:
   - Identify all files and components that will need changes
   - Understand the existing architecture and patterns
   - Find related code, tests, and dependencies
   - Note any potential challenges or edge cases

4. Create a comprehensive implementation plan that includes:
   - A clear summary of the approach
   - Step-by-step breakdown of changes needed
   - Files to be created or modified
   - Any database migrations required
   - Test coverage requirements

5. Format the refined issue with:
   - Original issue description preserved at the top
   - A "## Implementation Plan" section with the approach
   - A "## Checklist" section with actionable task items using `- [ ]` format

6. Update the issue on GitHub:
   `gh issue edit $1 --repo mjc-gh/pisces --body "REFINED_BODY"`

7. Report success and show a summary of the plan added to the issue.

**Important:**
- Preserve the original issue content; append the plan below it
- Keep checklist items specific and actionable
- Reference specific files and line numbers where helpful
- Follow the project conventions in AGENTS.md
