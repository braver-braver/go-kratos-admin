# Test Report: backend

- Command: `go test ./...` (run from `backend`)
- Result: **PASSED**
- Notes:
  - MinIO integration tests are gated by a reachability check to avoid failing when MinIO is not running locally.
  - Field mask tests use valid masks; jwt payload test compares client ID via getter to align with pointer field.
