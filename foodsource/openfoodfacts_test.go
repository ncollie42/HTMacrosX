package foodsource

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenFoodFactsLookupBarcode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": 1,
			"product": {
				"product_name_en": "Banana",
				"product_name": "Banana fallback",
				"nutriments": {
					"fat_100g": 0.3,
					"carbohydrates_100g": 22.8,
					"fiber_100g": 2.6,
					"proteins_100g": 1.1
				}
			}
		}`))
	}))
	defer server.Close()

	provider := &OpenFoodFactsProvider{Client: server.Client()}
	provider.Client.Transport = rewriteTransport(server.URL, provider.Client.Transport)

	got, err := provider.LookupBarcode("12345")
	if err != nil {
		t.Fatalf("LookupBarcode returned error: %v", err)
	}
	if got.Code != "12345" || got.Name != "Banana" {
		t.Fatalf("unexpected candidate: %+v", got)
	}
	if got.Fat != 0.3 || got.Carb != 22.8 || got.Fiber != 2.6 || got.Protein != 1.1 {
		t.Fatalf("unexpected macros: %+v", got)
	}
}

func TestOpenFoodFactsLookupBarcodeNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": 0}`))
	}))
	defer server.Close()

	provider := &OpenFoodFactsProvider{Client: server.Client()}
	provider.Client.Transport = rewriteTransport(server.URL, provider.Client.Transport)

	if _, err := provider.LookupBarcode("12345"); err != ErrNotFound {
		t.Fatalf("LookupBarcode err = %v, want ErrNotFound", err)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func rewriteTransport(baseURL string, fallback http.RoundTripper) http.RoundTripper {
	if fallback == nil {
		fallback = http.DefaultTransport
	}
	baseURL = strings.TrimPrefix(baseURL, "http://")
	baseURL = strings.TrimPrefix(baseURL, "https://")
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = baseURL
		return fallback.RoundTrip(req)
	})
}
