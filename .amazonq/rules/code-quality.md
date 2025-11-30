# Code Quality Rules

## Minimal Code Implementation
Write only the ABSOLUTE MINIMAL amount of code needed to address the requirement correctly.

**Principles**:
- Avoid verbose implementations
- No code that doesn't directly contribute to the solution
- No unnecessary abstractions
- No premature optimization
- No boilerplate unless required

**Example**:
```go
// ❌ Verbose
type Response struct {
    Success bool
    Data interface{}
    Error string
    Meta interface{}
    Timestamp time.Time
    RequestID string
    Version string
}

// ✅ Minimal
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

## Follow Project Conventions
- Use existing patterns (handlers, middleware, models)
- Match existing naming conventions
- Reuse existing utilities and helpers
- Don't reinvent the wheel

## Error Handling
- Use standardized response helpers from `internal/api/response.go`
- Log errors with context (tenant_id, user_id, IP)
- Return user-friendly error messages

## Database Queries
- Always use parameterized queries (prevent SQL injection)
- Add indexes for frequently queried columns
- Use pagination for large result sets
- Include proper error handling

## Security
- Validate all user inputs
- Check authentication/authorization
- Log security-relevant events to audit logs
- Never expose sensitive data in responses

---

**Quality over quantity. Simple over complex.**
