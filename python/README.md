# answeringservicecare

Official Python client and CLI for the [Answering Service Care](https://answeringservicecare.com) public API.

Answering Service Care (ASC) is a US answering service with live and AI receptionists, bilingual (English/Spanish) and available 24/7. The public API is read-only and needs no key: company details, services, current plan pricing, and site search.

- API reference: [openapi.json](https://answeringservicecare.com/openapi.json)
- Developer overview: [answeringservicecare.com/developers](https://answeringservicecare.com/developers/)
- JavaScript package: [`answeringservicecare` on npm](https://www.npmjs.com/package/answeringservicecare)

## Install

```bash
pip install answeringservicecare
```

Requires Python 3.9 or later. No dependencies.

## Use it

```python
from answeringservicecare import AnsweringServiceCare

asc = AnsweringServiceCare()

# Plans and prices in USD, optionally narrowed to one service or billing cycle
pricing = asc.get_pricing(service="answering service", billing_cycle="monthly")
for group in pricing["services"]:
    for plan in group["plans"]:
        print(f'{group["service"]} / {plan["name"]}: {plan["price"]} ({plan["included"]}), overage {plan["overage"]}')

services = asc.list_services()       # [{"name", "url", "markdownUrl", "description"}, ...]
company = asc.get_company()          # phone, email, address, founding date, links
results = asc.search("legal answering service", limit=5)
```

Responses are plain dicts with the API's own keys, typed with `TypedDict`s (`Pricing`, `Plan`, `Service`, `Company`, `SearchResult`). Every page on answeringservicecare.com is also available as markdown. Each service and search result carries a `markdownUrl`.

### Errors

Failed requests raise `AscApiError` with the HTTP `status`, a machine-readable `code`, and a `hint` saying what to change:

```python
from answeringservicecare import AscApiError

try:
    asc.search("a")
except AscApiError as error:
    print(error.code, error.hint)  # rest_invalid_param, Check the parameter constraints...
```

`status` is `0` and `code` is `network_error` when the request never got a response.

### Options

```python
AnsweringServiceCare(
    base_url="https://answeringservicecare.com/api/v1",  # default
    timeout=10.0,                                         # seconds, default
)
```

## CLI

```bash
answeringservicecare pricing --service "live chat" --billing annual
answeringservicecare services
answeringservicecare company
answeringservicecare search "medical answering service" --limit 5
```

Or without installing: `pipx run answeringservicecare pricing`. Every command prints JSON, so it pipes into `jq` and is easy for agents to read. `python -m answeringservicecare` works too.

## For AI agents

- MCP server with the same data as tools: `https://answeringservicecare.com/mcp` (no auth)
- Agent skills: [Answering-Service-Care/agent-skills](https://github.com/Answering-Service-Care/agent-skills)
- When to use ASC and how: [llms.md](https://answeringservicecare.com/llms.md)

## License

MIT
