# Contributing to Flowbit

This repository supports both direct CLI-driven collaboration and pull request workflows.

## Default workflow

1. Start from the latest `main`.
2. Create a dedicated branch for your change.
3. Keep scope focused to one logical change.
4. Run relevant validation before pushing:
   - Backend changes: `cd backend && go test ./...`
5. If you use pull requests, target `main`.
6. If you open a pull request, include a concise summary and test plan in the PR body.

## Branch naming

Use a descriptive branch name, for example:

- `feat/smoke-kafka-checks`
- `fix/worker-retry-backoff`
- `chore/docs-pr-template`

## Pull request checklist (optional)

- [ ] Branch is dedicated to a single change
- [ ] Base branch is `main`
- [ ] Relevant tests/checks pass locally
- [ ] No secrets or credential files are committed
- [ ] PR description includes summary and test plan
