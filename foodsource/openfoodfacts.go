package foodsource

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var ErrNotFound = errors.New("food source product not found")

const openFoodFactsBaseURL = "https://world.openfoodfacts.net"
const openFoodFactsUserAgent = "HTMacrosX/1.0 (+https://purple-cherry-7894.fly.dev)"

type Candidate struct {
	Code    string
	Name    string
	Fat     float64
	Carb    float64
	Fiber   float64
	Protein float64
}

type Provider interface {
	LookupBarcode(code string) (Candidate, error)
}

type OpenFoodFactsProvider struct {
	Client *http.Client
}

func (p *OpenFoodFactsProvider) LookupBarcode(code string) (Candidate, error) {
	resp, err := p.doRequest(openFoodFactsBaseURL + "/api/v2/product/" + url.PathEscape(code) + ".json")
	if err != nil {
		return Candidate{}, fmt.Errorf("lookup barcode: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Candidate{}, fmt.Errorf("lookup barcode: API returned %d", resp.StatusCode)
	}

	var result struct {
		Status  int `json:"status"`
		Product struct {
			ProductName   string `json:"product_name"`
			ProductNameEN string `json:"product_name_en"`
			Nutriments    struct {
				Fat100g   float64 `json:"fat_100g"`
				Carbs100g float64 `json:"carbohydrates_100g"`
				Fiber100g float64 `json:"fiber_100g"`
				Prot100g  float64 `json:"proteins_100g"`
			} `json:"nutriments"`
		} `json:"product"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Candidate{}, fmt.Errorf("lookup barcode: decode response: %w", err)
	}
	if result.Status != 1 {
		return Candidate{}, ErrNotFound
	}

	candidate := candidateFromProduct(code, result.Product.ProductNameEN, result.Product.ProductName, result.Product.Nutriments.Fat100g, result.Product.Nutriments.Carbs100g, result.Product.Nutriments.Fiber100g, result.Product.Nutriments.Prot100g)
	if candidate.Name == "" {
		candidate.Name = "Unknown (" + code + ")"
	}
	return candidate, nil
}

func (p *OpenFoodFactsProvider) client() *http.Client {
	if p != nil && p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}

func (p *OpenFoodFactsProvider) doRequest(rawURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", openFoodFactsUserAgent)
	return p.client().Do(req)
}

func candidateFromProduct(code string, preferredName string, fallbackName string, fat float64, carb float64, fiber float64, protein float64) Candidate {
	name := strings.TrimSpace(preferredName)
	if name == "" {
		name = strings.TrimSpace(fallbackName)
	}
	return Candidate{
		Code:    strings.TrimSpace(code),
		Name:    name,
		Fat:     fat,
		Carb:    carb,
		Fiber:   fiber,
		Protein: protein,
	}
}
