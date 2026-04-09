---
description: Stage changes, commit with semantic message, and push to main
model: openrouter/anthropic/claude-haiku-4.5
---

Stage all changes, write a semantic commit message, and push to main.

**Instructions:**

1. First, gather context by running these commands:
   - `git status` to see all changed files
   - `git diff --staged` and `git diff` to understand what changed
   - `git log --oneline -5` to see recent commit message style
   - `git branch --show-current` to confirm the current branch

2. Check if there's a GitHub issue being worked on by looking at recent conversation context or branch name. If an issue number is referenced, include it in the commit message.

3. Stage all relevant changes with `git add`. Do not stage files that contain secrets (.env, credentials, etc.).

4. Write a semantic commit message following this format:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `style:` for formatting changes
   - `refactor:` for code refactoring
   - `test:` for adding/updating tests
   - `chore:` for maintenance tasks

   If a GitHub issue is being addressed, reference it at the end: `fixes #123` or `closes #123`

5. Create the commit with `git commit -m "message"`

6. Push to main with `git push origin main`

7. Report the result including the commit hash and what was pushed.

**Important:**
- Do not include "Co-authored-by" or any AI attribution in the commit message
- Keep the commit message concise (under 72 characters for the subject line)
- If there are no changes to commit, inform the user
- If not on main branch, warn the user before pushing
