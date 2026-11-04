# Answering Service Care SDKs

Official clients and CLIs for the [Answering Service Care](https://answeringservicecare.com) public API.

Answering Service Care (ASC) is a US answering service with live and AI receptionists, bilingual (English/Spanish) and available 24/7. The public API is read-only and needs no key:

| Endpoint | What it returns |
|---|---|
| `GET /api/v1/company` | Company profile: phone, email, postal address, founding date, links |
| `GET /api/v1/services` | Every service ASC offers, with page and markdown URLs |
| `GET /api/v1/pricing` | Current plans and prices in USD, grouped by service |
| `GET /api/v1/search?q=` | Search across ASC's service, industry, location and blog pages |

Base URL: `https://answeringservicecare.com/api/v1`. Full description: [openapi.json](https://answeringservicecare.com/openapi.json).

## Packages

| Language | Package | Install | Source |
|---|---|---|---|
| JavaScript / TypeScript | [`answeringservicecare`](https://www.npmjs.com/package/answeringservicecare) | `npm install answeringservicecare` | [`js/`](js/) |
| Python | [`answeringservicecare`](https://pypi.org/project/answeringservicecare/) | `pip install answeringservicecare` | [`python/`](python/) |
| Go | [`github.com/Answering-Service-Care/sdk/go`](https://pkg.go.dev/github.com/Answering-Service-Care/sdk/go) | `go get github.com/Answering-Service-Care/sdk/go` | [`go/`](go/) |

All three include an `answeringservicecare` command:

```bash
npx answeringservicecare pricing --service "answering service" --billing monthly
pipx run answeringservicecare search "legal answering service" --limit 5
go run github.com/Answering-Service-Care/sdk/go/cmd/answeringservicecare@latest company
```

## Other ways in

- **MCP server** with the same data as tools: `https://answeringservicecare.com/mcp` (Streamable HTTP, no auth)
- **Agent skills and plugin**: [Answering-Service-Care/agent-skills](https://github.com/Answering-Service-Care/agent-skills)
- **Markdown**: every page is available as markdown; add `.md` to the path (homepage: `/index.md`)
- **Developer overview**: [answeringservicecare.com/developers](https://answeringservicecare.com/developers/)
- **Existing customers**: the authenticated customer API (messages, call recordings, to-dos) is documented at [answeringservicecare.com/docs/api-reference](https://answeringservicecare.com/docs/api-reference/)

## License

MIT
