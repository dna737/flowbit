# Agent notes

## Go standard library (Context7)

When writing, reviewing, or debugging Go in this repository, use the Context7 curated standard-library reference for up-to-date patterns and package behavior:

https://context7.com/golang/go/llms.txt?tokens=10000

The optional `tokens` query parameter caps how much text is returned when fetching that URL (useful for tool calls).

Primary Go code lives under `backend/`. After substantive changes, run `go test ./...` from `backend/`.
