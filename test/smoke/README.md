# Smoke test

End-to-end check of cch's non-interactive subcommands (`help`, `schema`,
`version`) run against the real binary via `go run`. Gated behind the
`integration` build tag so it does not run in the default unit-test suite.

Run it manually:

```bash
go test ./test/smoke -tags=integration -v
```
