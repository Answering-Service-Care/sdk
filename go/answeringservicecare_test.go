package answeringservicecare

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T) *Client {
	t.Helper()
	mux := http.NewServeMux()
	json := func(w http.ResponseWriter, status int, body string) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
	mux.HandleFunc("/api/v1/company", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "answeringservicecare-go/"+Version {
			t.Errorf("User-Agent = %q", got)
		}
		json(w, 200, `{"name":"Answering Service Care","telephone":"+1-800-000-0000","address":{"addressRegion":"FL"},"sameAs":["https://github.com/Answering-Service-Care"]}`)
	})
	mux.HandleFunc("/api/v1/services", func(w http.ResponseWriter, r *http.Request) {
		json(w, 200, `{"services":[{"name":"Answering Service","url":"https://answeringservicecare.com/services/answering-service/"}],"offerings":{}}`)
	})
	mux.HandleFunc("/api/v1/pricing", func(w http.ResponseWriter, r *http.Request) {
		json(w, 200, `{"currency":"USD","services":[
			{"service":"Answering Service","plans":[{"name":"Starter","billingCycle":"monthly","price":"$40"},{"name":"Starter","billingCycle":"annual","price":"$400"}]},
			{"service":"Live Chat","plans":[{"name":"Chat","billingCycle":"monthly","price":"$5"}]}]}`)
	})
	mux.HandleFunc("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query().Get("q"); q == "x" {
			json(w, 400, `{"code":"rest_invalid_param","message":"Invalid parameter(s): q","data":{"status":400,"params":{"q":"q must be at least 2 characters long."}},"hint":"Send 2-200 characters."}`)
			return
		}
		if r.URL.Query().Get("q") != "legal" || r.URL.Query().Get("limit") != "5" {
			t.Errorf("search query = %v", r.URL.Query())
		}
		json(w, 200, `{"query":"legal","results":[{"title":"Legal answering service","url":"https://answeringservicecare.com/industries/legal/"}],"nextCursor":null}`)
	})
	mux.HandleFunc("/api/v1/blocked", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(403)
		_, _ = w.Write([]byte("<html>Just a moment...</html>"))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := New()
	client.BaseURL = server.URL + "/api/v1/"
	return client
}

func TestGetCompany(t *testing.T) {
	company, err := testServer(t).GetCompany(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if company.Name != "Answering Service Care" || company.Address.AddressRegion != "FL" || company.SameAs[0] != "https://github.com/Answering-Service-Care" {
		t.Errorf("company = %+v", company)
	}
}

func TestListServices(t *testing.T) {
	services, err := testServer(t).ListServices(context.Background())
	if err != nil || len(services) != 1 || services[0].Name != "Answering Service" {
		t.Fatalf("services = %+v, err = %v", services, err)
	}
}

func TestGetPricingFilters(t *testing.T) {
	client := testServer(t)
	all, err := client.GetPricing(context.Background(), PricingFilter{})
	if err != nil || len(all.Services) != 2 || all.Currency != "USD" {
		t.Fatalf("unfiltered = %+v, err = %v", all, err)
	}
	monthly, _ := client.GetPricing(context.Background(), PricingFilter{Service: "ANSWERING", BillingCycle: "monthly"})
	if len(monthly.Services) != 1 || len(monthly.Services[0].Plans) != 1 || monthly.Services[0].Plans[0].Price != "$40" {
		t.Errorf("service+monthly = %+v", monthly.Services)
	}
	annual, _ := client.GetPricing(context.Background(), PricingFilter{BillingCycle: "annual"})
	if len(annual.Services) != 1 || annual.Services[0].Service != "Answering Service" {
		t.Errorf("annual drops services with no annual plan: %+v", annual.Services)
	}
}

func TestSearch(t *testing.T) {
	results, err := testServer(t).Search(context.Background(), "  legal ", 5)
	if err != nil || len(results) != 1 || results[0].Title != "Legal answering service" {
		t.Fatalf("results = %+v, err = %v", results, err)
	}
}

func TestAPIError(t *testing.T) {
	_, err := testServer(t).Search(context.Background(), "x", 0)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v", err)
	}
	if apiErr.Status != 400 || apiErr.Code != "rest_invalid_param" || apiErr.Hint != "Send 2-200 characters." || apiErr.DocumentationURL != docsURL {
		t.Errorf("apiErr = %+v", apiErr)
	}
	if apiErr.Message != "Invalid parameter(s): q (q must be at least 2 characters long.)" {
		t.Errorf("message = %q", apiErr.Message)
	}
}

func TestNonJSONAndUnknownRoute(t *testing.T) {
	client := testServer(t)
	var apiErr *APIError
	if err := client.get(context.Background(), "/blocked", &struct{}{}); !errors.As(err, &apiErr) || apiErr.Status != 403 || apiErr.Code != "unexpected_response" || apiErr.Hint == "" {
		t.Errorf("blocked: %+v", apiErr)
	}
	if err := client.get(context.Background(), "/nope", &struct{}{}); !errors.As(err, &apiErr) || apiErr.Status != 404 {
		t.Errorf("unknown route: %+v", apiErr)
	}
}

func TestNetworkError(t *testing.T) {
	client := New()
	client.BaseURL = "http://127.0.0.1:1/api/v1"
	_, err := client.GetCompany(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 0 || apiErr.Code != "network_error" {
		t.Errorf("err = %+v", err)
	}
}
