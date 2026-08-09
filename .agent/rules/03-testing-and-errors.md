# Testing, Errors & Response Envelope Specification

> **Rule ID:** `03-testing-and-errors`  
> **Applicable Globs:** `internal/**/*_test.go`, `internal/models/response.go`, `internal/handler/**`, `internal/middleware/**`

---

## 1. Response Envelope Standard

All HTTP responses returned by the Gin server must strictly adhere to the standard envelope format defined in [internal/models/response.go](file:///internal/models/response.go):

### A. Success Envelope (`SendSuccess`)
```go
models.SendSuccess(c, http.StatusOK, gin.H{"user": userProfile}, "User fetched successfully")
```
Output:
```json
{
  "success": true,
  "data": {
    "user": { ... }
  },
  "message": "User fetched successfully"
}
```

### B. Error Envelope (`SendError`)
```go
models.SendError(c, http.StatusBadRequest, "INVALID_INPUT", "Workout date cannot be in the future", []string{"date: 2099-01-01 exceeds current date"})
```
Output:
```json
{
  "success": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "Workout date cannot be in the future",
    "details": ["date: 2099-01-01 exceeds current date"]
  },
  "timestamp": "2026-08-09T07:20:00Z"
}
```

---

## 2. HTTP Status Code Guidelines

| Status Code | Error Code | When to Use |
| :--- | :--- | :--- |
| **200 OK** | - | Successful fetch, update, or deletion |
| **201 Created** | - | Successful resource creation (e.g. user bootstrap or log creation) |
| **400 Bad Request** | `INVALID_INPUT` / `BAD_REQUEST` | Malformed JSON body, invalid date syntax, bad query parameters |
| **401 Unauthorized**| `UNAUTHORIZED` | Missing, invalid, or expired Supabase JWT token |
| **403 Forbidden** | `FORBIDDEN` | Authenticated user attempting to modify another user's resources |
| **404 Not Found** | `NOT_FOUND` | Requested entity (user profile, gym log, plan) does not exist |
| **409 Conflict** | `CONFLICT` | Duplicate entry or constraint violation |
| **500 Internal Error** | `INTERNAL_ERROR` | Database connection drops, query failure, unexpected runtime error |

---

## 3. Unit & Integration Testing Standards

1. **Table-Driven Tests:**
   - Use table-driven test patterns with clear `name`, input parameters, and expected outcomes.
   - Example reference: [internal/service/stats_calculator_test.go](file:///internal/service/stats_calculator_test.go).

2. **Test Command Execution:**
   - Run all package tests: `go test -v ./...`
   - Run package-specific tests: `go test -v ./internal/service`
   - Check test coverage: `go test -cover ./...`

3. **Mocking & Isolation:**
   - Unit tests for services must mock repository interfaces from `interfaces.go` rather than requiring a live PostgreSQL instance.

---

## 4. The 3-Attempt Circuit Breaker Rule

If a bug, compile error, test failure, or database issue persists after **3 consecutive fix attempts**:
1. **STOP IMMEDIATELY.**
2. Revert the code changes back to the last known working Git commit.
3. Open the target feature's `STATE.md` and log the exact error message, reproduction steps, and root cause hypothesis in the **Execution Log**.
4. Ask the human engineer for architectural guidance.
