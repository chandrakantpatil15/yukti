# Development Workflow Rules

## Rule 1: Always Ask for Files First
Always ask me to open relevant files before answering. Never guess code.

**Example**:
- ❌ "I'll update the handler to add pagination..."
- ✅ "Please open `internal/api/handlers/resources.go` so I can see the current implementation."

## Rule 2: Admit When Unsure
If unsure or missing context, say: "Need file/clarification" instead of inventing.

**Example**:
- ❌ Making assumptions about database schema
- ✅ "Need file/clarification: Which table stores the resource metrics?"

## Rule 3: Work in Small Steps
Work in small steps. After each step, WAIT for my confirmation.

**Example**:
- ❌ Implementing 5 endpoints at once
- ✅ "I've added the pagination logic. Should I proceed with adding the filter parameters?"

## Rule 4: Show Only Changed Lines
When editing code, show ONLY the changed lines, not full files.

**Example**:
- ❌ Showing entire 500-line file
- ✅ Showing only the 10 lines that changed with context

## Rule 5: Follow Existing Patterns
Follow existing project patterns; do not introduce new structure unless requested.

**Example**:
- ❌ Creating new error handling pattern when one exists
- ✅ Using existing `api.Error()` helper from `internal/api/response.go`

## Rule 6: Ask When Stuck
If stuck, stop and ask: "What is the next file or detail?"

**Example**:
- ❌ Continuing with incomplete information
- ✅ "What is the next file or detail? I need to know the JWT claims structure."

---

**These rules ensure efficient, accurate, and collaborative development.**
