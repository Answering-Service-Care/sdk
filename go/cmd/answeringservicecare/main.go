// Command answeringservicecare prints read-only facts about Answering Service
// Care (ASC), a US answering service with live and AI receptionists, as JSON.
// No API key needed.
//
//	answeringservicecare pricing --service "answering service" --billing monthly
//	answeringservicecare search --limit 5 legal answering service
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	asc "github.com/Answering-Service-Care/sdk/go"
)

const usage = `answeringservicecare: read-only facts about Answering Service Care (ASC),
a US answering service with live and AI receptionists. No API key needed.
Output is JSON.

usage: answeringservicecare [--base-url URL] <command> [flags]

commands:
  company                                   Company profile and contact details
  services                                  Services ASC offers
  pricing [--service S] [--billing monthly|annual]
                                            Current plans and prices
  search [--limit N] <words...>             Search ASC's pages and articles

examples:
  answeringservicecare pricing --service "answering service" --billing monthly
  answeringservicecare search --limit 5 legal answering service

docs: https://answeringservicecare.com/developers/
`

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	global := flag.NewFlagSet("answeringservicecare", flag.ContinueOnError)
	global.SetOutput(stderr)
	global.Usage = func() { fmt.Fprint(stderr, usage) }
	baseURL := global.String("base-url", asc.DefaultBaseURL, "API base URL")
	version := global.Bool("version", false, "print the version")
	if err := global.Parse(args); err != nil {
		return 2
	}
	if *version {
		fmt.Fprintln(stdout, asc.Version)
		return 0
	}
	if global.NArg() == 0 {
		global.Usage()
		return 1
	}

	client := asc.New()
	client.BaseURL = *baseURL
	command, rest := global.Arg(0), global.Args()[1:]

	var data any
	var err error
	switch command {
	case "company":
		data, err = client.GetCompany(ctx)
	case "services":
		data, err = client.ListServices(ctx)
	case "pricing":
		flags := flag.NewFlagSet("pricing", flag.ContinueOnError)
		flags.SetOutput(stderr)
		service := flags.String("service", "", `only this service, e.g. "live chat"`)
		billing := flags.String("billing", "", "only monthly or only annual plans")
		if flags.Parse(rest) != nil {
			return 2
		}
		if *billing != "" && *billing != "monthly" && *billing != "annual" {
			fmt.Fprintln(stderr, "--billing must be monthly or annual")
			return 2
		}
		data, err = client.GetPricing(ctx, asc.PricingFilter{Service: *service, BillingCycle: *billing})
	case "search":
		flags := flag.NewFlagSet("search", flag.ContinueOnError)
		flags.SetOutput(stderr)
		limit := flags.Int("limit", 0, "maximum results, 1-20 (default 10)")
		if flags.Parse(rest) != nil {
			return 2
		}
		if flags.NArg() == 0 {
			fmt.Fprintln(stderr, "search needs words to search for")
			return 2
		}
		data, err = client.Search(ctx, strings.Join(flags.Args(), " "), *limit)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", command)
		global.Usage()
		return 2
	}

	if err != nil {
		var apiErr *asc.APIError
		if errors.As(err, &apiErr) {
			status := ""
			if apiErr.Status != 0 {
				status = fmt.Sprintf(", HTTP %d", apiErr.Status)
			}
			fmt.Fprintf(stderr, "%s (%s%s)\n", apiErr.Error(), apiErr.Code, status)
		} else {
			fmt.Fprintln(stderr, err)
		}
		return 1
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
