# AGENTS.md

Official JavaScript and Python clients for the Answering Service Care public API (`https://answeringservicecare.com/api/v1`, described by `/openapi.json`). Both packages are published as `answeringservicecare`, have no runtime dependencies, and expose the same four calls plus an `answeringservicecare` CLI that prints JSON.

## Layout

- `js/`: TypeScript source in `src/` (`index.ts` client, `cli.ts` commands, `bin.ts` executable), vitest tests in `test/`.
- `python/`: package in `src/answeringservicecare/` (`client.py`, `types.py`, `cli.py`), pytest tests in `tests/` that run against a local HTTP server.

## Commands

```bash
# JavaScript (Node 20+)
cd js && npm ci && npm run typecheck && npm test && npm run build

# Python (3.9+)
cd python && python -m venv .venv && . .venv/bin/activate
pip install -e . pytest mypy && pytest && mypy --strict src
```

Try either CLI against a local copy of the site with `--base-url http://localhost:8080/api/v1`.

## Rules

- Keep the two packages in step: same methods, same filter behavior, same error fields (`status`, `code`, `hint`, documentation URL), same CLI commands and flags.
- The API is read-only. Don't add methods that write, and don't add an API key option; none exists.
- Response types mirror `components.schemas` in the site's `/openapi.json`. Update both packages when it changes.
- No runtime dependencies: `fetch` in JavaScript, `urllib` in Python.
- Bump the version in `js/package.json` and `js/src/index.ts` (`VERSION`), or `python/pyproject.toml` and `python/src/answeringservicecare/client.py` (`__version__`), together. Tests check they match.
