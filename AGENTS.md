# AGENTS.md

Official JavaScript, Python, and Go clients for the Answering Service Care public API (`https://answeringservicecare.com/api/v1`, described by `/openapi.json`). The JavaScript and Python packages are published as `answeringservicecare` and the Go module as `github.com/Answering-Service-Care/sdk/go`; all have no runtime dependencies and expose the same four calls plus an `answeringservicecare` CLI that prints JSON.

## Layout

- `js/`: TypeScript source in `src/` (`index.ts` client, `cli.ts` commands, `bin.ts` executable), vitest tests in `test/`.
- `python/`: package in `src/answeringservicecare/` (`client.py`, `types.py`, `cli.py`), pytest tests in `tests/` that run against a local HTTP server.
- `go/`: module `github.com/Answering-Service-Care/sdk/go` (`answeringservicecare.go` client and types), CLI in `cmd/answeringservicecare/`, tests against `httptest` servers.

## Commands

```bash
# JavaScript (Node 20+)
cd js && npm ci && npm run typecheck && npm test && npm run build

# Python (3.9+)
cd python && python -m venv .venv && . .venv/bin/activate
pip install -e . pytest mypy && pytest && mypy --strict src

# Go (1.21+)
cd go && go vet ./... && go test ./...
```

Try either CLI against a local copy of the site with `--base-url http://localhost:8080/api/v1`.

## Rules

- Keep the three packages in step: same methods, same filter behavior, same error fields (`status`, `code`, `hint`, documentation URL), same CLI commands and flags.
- The API is read-only. Don't add methods that write, and don't add an API key option; none exists.
- Response types mirror `components.schemas` in the site's `/openapi.json`. Update both packages when it changes.
- No runtime dependencies: `fetch` in JavaScript, `urllib` in Python, `net/http` in Go.
- Bump the version in `js/package.json` and `js/src/index.ts` (`VERSION`), or `python/pyproject.toml` and `python/src/answeringservicecare/client.py` (`__version__`), or `go/answeringservicecare.go` (`Version`) and the `go/vX.Y.Z` release tag, together. Tests check they match.
