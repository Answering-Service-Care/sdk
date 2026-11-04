# answeringservicecare (Go)

Official Go client and CLI for the [Answering Service Care](https://answeringservicecare.com) public API: read-only company details, services, plan pricing, and site search. No API key needed. Homepage: [answeringservicecare.com/developers](https://answeringservicecare.com/developers/).

```bash
go get github.com/Answering-Service-Care/sdk/go
go install github.com/Answering-Service-Care/sdk/go/cmd/answeringservicecare@latest
```

```go
package main

import (
	"context"
	"fmt"

	asc "github.com/Answering-Service-Care/sdk/go"
)

func main() {
	client := asc.New()
	pricing, err := client.GetPricing(context.Background(), asc.PricingFilter{Service: "answering service", BillingCycle: "monthly"})
	if err != nil {
		panic(err) // *asc.APIError carries Status, Code, Hint, and DocumentationURL
	}
	for _, group := range pricing.Services {
		for _, plan := range group.Plans {
			fmt.Println(group.Service, plan.Name, plan.Price, plan.Included)
		}
	}
}
```

| Method | Endpoint |
|---|---|
| `GetCompany(ctx)` | `GET /api/v1/company` |
| `ListServices(ctx)` | `GET /api/v1/services` |
| `GetPricing(ctx, PricingFilter{Service, BillingCycle})` | `GET /api/v1/pricing`, filtered like the JS and Python clients |
| `Search(ctx, query, limit)` | `GET /api/v1/search` |

The CLI prints JSON:

```bash
answeringservicecare pricing --service "answering service" --billing monthly
answeringservicecare search --limit 5 legal answering service
answeringservicecare --base-url https://answeringservicecare.com/api/sandbox/v1 pricing   # fixed test data
```

No dependencies beyond the standard library. MIT licensed.
