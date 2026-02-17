---
name: update-deps
description: Updates Go module dependencies with full verification pipeline (get, tidy, fmt, lint, test).
user_invocable: true
---

# Update Go Dependencies

This skill automates the process of updating Go dependencies in the order service. It handles the full lifecycle: checking for updates, applying them, and running the verification pipeline.

## Step 0: Pre-flight Checks

Before doing anything, verify the environment:

1. **Check current branch** — run `git branch --show-current`. If on `main`, STOP and ask the user to switch to a feature branch first.
2. **Check for uncommitted changes** — run `git status --porcelain`. If there are uncommitted changes, WARN the user and ask whether to proceed or abort.

## Step 1: Show Available Updates

Run:
```bash
go list -m -u all 2>/dev/null | grep '\[.*\]'
```

This shows all modules with available updates. Present the list to the user clearly, highlighting:
- **Critical packages** (require extra caution): `connectrpc.com/connect`, `google.golang.org/protobuf`, `github.com/jmoiron/sqlx`, `github.com/lib/pq`
- **Standard packages**: everything else

**Note**: `github.com/demo/contracts` uses a local replace directive and cannot be updated via `go get`. Skip it in the update list.

Ask the user which update strategy to use:
- **All updates** — update everything to latest
- **Patch only** — only patch-level updates (safest)
- **Specific packages** — user picks which packages to update

## Step 2: Update Dependencies

Based on the user's choice, run the appropriate commands:

### All updates
```bash
go get -u ./...
go mod tidy
```

### Patch only
```bash
go get -u=patch ./...
go mod tidy
```

### Specific packages
```bash
go get package@version
go mod tidy
```

For specific packages, the user provides the list. Use `go get package@latest` or `go get package@vX.Y.Z` as requested.

## Step 3: Show Changes

Show the user what changed:
```bash
git diff go.mod go.sum
```

Summarize the key changes: which packages were updated, from which version to which version. If critical packages were updated, explicitly call them out.

## Step 4: Verification Pipeline

Run the full verification pipeline in order. Stop on first failure.

1. **Format code** — `task fmt`
2. **Lint** — `task lint`
3. **Build** — `task build`
4. **Run tests** — `task test`

If all steps pass, report success.

## Step 5: Handle Failures

If any verification step fails:

1. Show the error output clearly
2. Offer the user three options:
   - **Fix** — attempt to fix the issue (compilation errors, lint issues, etc.)
   - **Rollback problematic packages** — revert specific packages that caused the issue using `go get package@old-version` and re-run the pipeline
   - **Abort** — revert all changes with `git checkout go.mod go.sum`

## Step 6: Commit (Only When User Requests)

Do NOT auto-commit. When the user asks to commit:

1. Stage the changed files: `go.mod`, `go.sum`, and any other changed files
2. Create a commit message:

```
update go dependencies

Updated packages:
- package/name: v1.2.3 -> v1.4.0
- other/package: v0.5.0 -> v0.6.1
```

## Important Notes

### Local replace directives
The module `github.com/demo/contracts` is replaced with a local path (`../contracts`). It cannot be updated via `go get` and should be excluded from the update list.

### Critical packages requiring extra caution
These packages are tightly coupled to core functionality. Updates may require code changes:
- `connectrpc.com/connect` — Connect RPC framework, changes can affect handler signatures
- `google.golang.org/protobuf` — protobuf runtime, must be compatible with generated code in contracts
- `github.com/jmoiron/sqlx` — database access layer, changes may affect query execution
- `github.com/lib/pq` — PostgreSQL driver, changes may affect database connectivity
