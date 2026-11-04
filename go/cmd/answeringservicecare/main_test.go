package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	asc "github.com/Answering-Service-Care/sdk/go"
)

func TestPricingCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"services":[{"service":"Live Chat","plans":[{"name":"Chat","billingCycle":"monthly","price":"$5"}]},{"service":"Answering Service","plans":[{"name":"Starter","billingCycle":"annual","price":"$400"}]}]}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"--base-url", server.URL, "pricing", "--service", "chat", "--billing", "monthly"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	var pricing asc.Pricing
	if err := json.Unmarshal(stdout.Bytes(), &pricing); err != nil || len(pricing.Services) != 1 || pricing.Services[0].Service != "Live Chat" {
		t.Errorf("output = %s, err = %v", stdout.String(), err)
	}
}

func TestUsageErrors(t *testing.T) {
	cases := map[string][]string{
		"no command":     {},
		"unknown":        {"bogus"},
		"bad billing":    {"pricing", "--billing", "weekly"},
		"search no args": {"search"},
	}
	for name, args := range cases {
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, &stdout, &stderr); code == 0 {
			t.Errorf("%s: exit 0", name)
		}
	}
	var stdout, stderr bytes.Buffer
	if run(context.Background(), []string{"--version"}, &stdout, &stderr) != 0 || strings.TrimSpace(stdout.String()) != asc.Version {
		t.Errorf("--version printed %q", stdout.String())
	}
}
