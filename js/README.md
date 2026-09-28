# answeringservicecare

Official JavaScript/TypeScript client and CLI for the [Answering Service Care](https://answeringservicecare.com) public API.

Answering Service Care (ASC) is a US answering service with live and AI receptionists, bilingual (English/Spanish) and available 24/7. The public API is read-only and needs no key: company details, services, current plan pricing, and site search.

- API reference: [openapi.json](https://answeringservicecare.com/openapi.json)
- Developer overview: [answeringservicecare.com/developers](https://answeringservicecare.com/developers/)
- Python package: [`answeringservicecare` on PyPI](https://pypi.org/project/answeringservicecare/)

## Install

```bash
npm install answeringservicecare
```

Requires Node.js 20 or later. It also works in browsers, Deno, Bun, and Cloudflare Workers, anywhere with `fetch`. No dependencies.

## Use it

```ts
import { AnsweringServiceCare } from "answeringservicecare";

const asc = new AnsweringServiceCare();

// Plans and prices in USD, optionally narrowed to one service or billing cycle
const pricing = await asc.getPricing({ service: "answering service", billingCycle: "monthly" });
for (const { service, plans } of pricing.services) {
  for (const plan of plans) {
    console.log(`${service} / ${plan.name}: ${plan.price} (${plan.included}), overage ${plan.overage}`);
  }
}

const services = await asc.listServices();        // [{ name, url, markdownUrl, description }, ...]
const company = await asc.getCompany();           // phone, email, address, founding date, links
const results = await asc.search("legal answering service", { limit: 5 });
```

Every page on answeringservicecare.com is also available as markdown. Each service and search result carries a `markdownUrl`.

### Errors

Failed requests throw `AscApiError` with the HTTP `status`, a machine-readable `code`, and a `hint` saying what to change:

```ts
import { AscApiError } from "answeringservicecare";

try {
  await asc.search("a");
} catch (error) {
  if (error instanceof AscApiError) {
    console.error(error.code, error.hint); // rest_invalid_param, Check the parameter constraints...
  }
}
```

`status` is `0` and `code` is `network_error` when the request never got a response.

### Options

```ts
new AnsweringServiceCare({
  baseUrl: "https://answeringservicecare.com/api/v1", // default
  timeoutMs: 10_000,                                 // default
  fetch: customFetch,                                // defaults to the global fetch
});
```

## CLI

```bash
npx answeringservicecare pricing --service "live chat" --billing annual
npx answeringservicecare services
npx answeringservicecare company
npx answeringservicecare search "medical answering service" --limit 5
```

Every command prints JSON, so it pipes into `jq` and is easy for agents to read. Run `npx answeringservicecare --help` for all options.

## For AI agents

- MCP server with the same data as tools: `https://answeringservicecare.com/mcp` (no auth)
- Agent skills: [Answering-Service-Care/agent-skills](https://github.com/Answering-Service-Care/agent-skills)
- When to use ASC and how: [llms.md](https://answeringservicecare.com/llms.md)

## License

MIT
