// Package answeringservicecare is the official Go client for the Answering
// Service Care (ASC) public API at https://answeringservicecare.com/api/v1:
// read-only company details, services, plan pricing, and site search. ASC is
// a US answering service with live and AI receptionists. No API key is
// needed. The API is described by https://answeringservicecare.com/openapi.json
// and documented at https://answeringservicecare.com/developers/.
package answeringservicecare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Version of this package; sent in the User-Agent header.
const Version = "1.0.0"

// DefaultBaseURL is the production API.
const DefaultBaseURL = "https://answeringservicecare.com/api/v1"

const docsURL = "https://answeringservicecare.com/developers/"

// Client calls the public API. The zero value is not usable; use New.
type Client struct {
	// BaseURL is the API base URL, without a trailing slash.
	BaseURL string
	// HTTPClient sends the requests.
	HTTPClient *http.Client
}

// New returns a Client for DefaultBaseURL with a 10-second timeout.
func New() *Client {
	return &Client{BaseURL: DefaultBaseURL, HTTPClient: &http.Client{Timeout: 10 * time.Second}}
}

// APIError is an error response from the API, or a response that wasn't the
// API's JSON.
type APIError struct {
	Message string
	// Status is the HTTP status, or 0 when the request never got a response.
	Status int
	// Code is machine-readable, e.g. "rest_no_route" or "rest_invalid_param".
	Code string
	// Hint says what to change to make the request succeed.
	Hint             string
	DocumentationURL string
}

func (e *APIError) Error() string {
	if e.Hint != "" {
		return e.Message + " " + e.Hint
	}
	return e.Message
}

// PostalAddress is a schema.org PostalAddress.
type PostalAddress struct {
	StreetAddress   string `json:"streetAddress,omitempty"`
	AddressLocality string `json:"addressLocality,omitempty"`
	AddressRegion   string `json:"addressRegion,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	AddressCountry  string `json:"addressCountry,omitempty"`
}

// Company is the company profile.
type Company struct {
	Name              string            `json:"name,omitempty"`
	URL               string            `json:"url,omitempty"`
	LegalName         string            `json:"legalName,omitempty"`
	AlternateName     string            `json:"alternateName,omitempty"`
	Description       string            `json:"description,omitempty"`
	Logo              string            `json:"logo,omitempty"`
	FoundingDate      string            `json:"foundingDate,omitempty"`
	NumberOfEmployees string            `json:"numberOfEmployees,omitempty"`
	Telephone         string            `json:"telephone,omitempty"`
	Email             string            `json:"email,omitempty"`
	Address           *PostalAddress    `json:"address,omitempty"`
	AreaServed        string            `json:"areaServed,omitempty"`
	SameAs            []string          `json:"sameAs,omitempty"`
	Links             map[string]string `json:"links,omitempty"`
}

// Service is one service ASC offers.
type Service struct {
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	Slug        string `json:"slug,omitempty"`
	MarkdownURL string `json:"markdownUrl,omitempty"`
	Description string `json:"description,omitempty"`
}

// Plan is one priced plan.
type Plan struct {
	Name         string `json:"name,omitempty"`
	BillingCycle string `json:"billingCycle,omitempty"`
	// Price as displayed, e.g. "$169".
	Price         string  `json:"price,omitempty"`
	PriceValue    float64 `json:"priceValue,omitempty"`
	PriceCurrency string  `json:"priceCurrency,omitempty"`
	Term          string  `json:"term,omitempty"`
	// Included is the usage included, e.g. "100 Minutes".
	Included string `json:"included,omitempty"`
	Billing  string `json:"billing,omitempty"`
	// Overage is the rate beyond the included usage.
	Overage     string   `json:"overage,omitempty"`
	Features    []string `json:"features,omitempty"`
	MostPopular bool     `json:"mostPopular,omitempty"`
	URL         string   `json:"url,omitempty"`
}

// ServicePlans groups the plans for one service.
type ServicePlans struct {
	Service string `json:"service"`
	Plans   []Plan `json:"plans"`
}

// Pricing is the current plan catalog in USD.
type Pricing struct {
	Heading      string         `json:"heading,omitempty"`
	Summary      string         `json:"summary,omitempty"`
	Notes        []string       `json:"notes,omitempty"`
	Currency     string         `json:"currency,omitempty"`
	URL          string         `json:"url,omitempty"`
	TermsURL     string         `json:"termsUrl,omitempty"`
	LastModified string         `json:"lastModified,omitempty"`
	Services     []ServicePlans `json:"services"`
}

// SearchResult is one page or article matching a search.
type SearchResult struct {
	Title        string `json:"title,omitempty"`
	URL          string `json:"url,omitempty"`
	MarkdownURL  string `json:"markdownUrl,omitempty"`
	Type         string `json:"type,omitempty"`
	Description  string `json:"description,omitempty"`
	LastModified string `json:"lastModified,omitempty"`
}

// PricingFilter narrows GetPricing. Empty fields don't filter.
type PricingFilter struct {
	// Service keeps only services whose name contains it (case-insensitive),
	// e.g. "answering service" or "live chat".
	Service string
	// BillingCycle keeps only "monthly" or only "annual" plans.
	BillingCycle string
}

// GetCompany returns the company profile: description, phone, email, postal
// address, founding date, social profiles, and key links.
func (c *Client) GetCompany(ctx context.Context) (*Company, error) {
	var company Company
	if err := c.get(ctx, "/company", &company); err != nil {
		return nil, err
	}
	return &company, nil
}

// ListServices returns every service ASC offers, with a description, page
// URL, and markdown URL.
func (c *Client) ListServices(ctx context.Context) ([]Service, error) {
	var body struct {
		Services []Service `json:"services"`
	}
	if err := c.get(ctx, "/services", &body); err != nil {
		return nil, err
	}
	return body.Services, nil
}

// GetPricing returns current plans and prices in USD, grouped by service.
func (c *Client) GetPricing(ctx context.Context, filter PricingFilter) (*Pricing, error) {
	var pricing Pricing
	if err := c.get(ctx, "/pricing", &pricing); err != nil {
		return nil, err
	}
	needle := strings.ToLower(filter.Service)
	services := make([]ServicePlans, 0, len(pricing.Services))
	for _, group := range pricing.Services {
		if needle != "" && !strings.Contains(strings.ToLower(group.Service), needle) {
			continue
		}
		if filter.BillingCycle != "" {
			plans := make([]Plan, 0, len(group.Plans))
			for _, plan := range group.Plans {
				if plan.BillingCycle == filter.BillingCycle {
					plans = append(plans, plan)
				}
			}
			if len(plans) == 0 {
				continue
			}
			group.Plans = plans
		}
		services = append(services, group)
	}
	pricing.Services = services
	return &pricing, nil
}

// Search runs a full-text search over ASC's service, industry, location, and
// blog pages. The query is 2-200 characters; limit is 1-20, or 0 for the
// API's default of 10.
func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	params := url.Values{"q": {strings.TrimSpace(query)}}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	var body struct {
		Results []SearchResult `json:"results"`
	}
	if err := c.get(ctx, "/search?"+params.Encode(), &body); err != nil {
		return nil, err
	}
	return body.Results, nil
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	target := strings.TrimRight(c.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "answeringservicecare-go/"+Version)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return &APIError{
			Message:          fmt.Sprintf("Could not reach %s: %v.", target, err),
			Code:             "network_error",
			Hint:             "Check the connection and try again.",
			DocumentationURL: docsURL,
		}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{Message: fmt.Sprintf("Could not read the response from %s: %v.", target, err), Status: resp.StatusCode, Code: "network_error", DocumentationURL: docsURL}
	}

	isJSON := strings.Contains(resp.Header.Get("Content-Type"), "json")
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && isJSON && json.Unmarshal(raw, out) == nil {
		return nil
	}

	var body struct {
		Code             string `json:"code"`
		Message          string `json:"message"`
		Hint             string `json:"hint"`
		DocumentationURL string `json:"documentation_url"`
		Data             struct {
			Params map[string]string `json:"params"`
		} `json:"data"`
	}
	if isJSON {
		_ = json.Unmarshal(raw, &body)
	}
	apiErr := &APIError{Message: body.Message, Status: resp.StatusCode, Code: body.Code, Hint: body.Hint, DocumentationURL: body.DocumentationURL}
	if apiErr.Message == "" {
		apiErr.Message = fmt.Sprintf("Unexpected %d response from %s.", resp.StatusCode, target)
	}
	if len(body.Data.Params) > 0 {
		details := make([]string, 0, len(body.Data.Params))
		for _, detail := range body.Data.Params {
			details = append(details, detail)
		}
		apiErr.Message += " (" + strings.Join(details, " ") + ")"
	}
	if apiErr.Code == "" {
		apiErr.Code = "unexpected_response"
	}
	if apiErr.Hint == "" && !isJSON {
		apiErr.Hint = "The response wasn't JSON; the request may have been blocked. Try again shortly."
	}
	if apiErr.DocumentationURL == "" {
		apiErr.DocumentationURL = docsURL
	}
	return apiErr
}
