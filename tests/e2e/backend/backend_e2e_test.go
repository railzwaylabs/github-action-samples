//go:build e2e

package backend_e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	baseURL    string
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

type order struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Status     int8   `json:"status"`
	Items      []struct {
		ID        string `json:"id"`
		ProductID string `json:"product_id"`
		Quantity  int64  `json:"quantity"`
		Price     int64  `json:"price"`
		Subtotal  int64  `json:"subtotal"`
	} `json:"items"`
}

func TestMain(m *testing.M) {
	baseURL = strings.TrimRight(os.Getenv("E2E_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	if err := waitUntilReady(30 * time.Second); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestHealthAndReadiness(t *testing.T) {
	for _, path := range []string{"/health", "/ready"} {
		response := request(t, http.MethodGet, path, nil)
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET %s: expected 200, got %d: %s", path, response.StatusCode, readBody(response.Body))
		}
	}
}

func TestOrderLifecycleAndFinancialCalculation(t *testing.T) {
	customerID := fmt.Sprintf("backend-e2e-%d", time.Now().UnixNano())
	payload := map[string]any{
		"customer_id": customerID,
		"items": []map[string]any{
			{"product_id": "growth-plan", "quantity": 2, "price": 1250},
			{"product_id": "support-addon", "quantity": 1, "price": 1499},
		},
	}
	createdResponse := request(t, http.MethodPost, "/orders", payload)
	defer createdResponse.Body.Close()
	if createdResponse.StatusCode != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", createdResponse.StatusCode, readBody(createdResponse.Body))
	}
	var created order
	decodeJSON(t, createdResponse.Body, &created)
	if created.ID == "" || created.CustomerID != customerID || created.Status != 1 {
		t.Fatalf("unexpected created order: %+v", created)
	}
	if len(created.Items) != 2 || created.Items[0].Subtotal != 2500 || created.Items[1].Subtotal != 1499 {
		t.Fatalf("financial invariant failed, expected subtotals [2500,1499]: %+v", created.Items)
	}

	fetchedResponse := request(t, http.MethodGet, "/orders/"+created.ID, nil)
	defer fetchedResponse.Body.Close()
	if fetchedResponse.StatusCode != http.StatusOK {
		t.Fatalf("get: expected 200, got %d: %s", fetchedResponse.StatusCode, readBody(fetchedResponse.Body))
	}
	var fetched order
	decodeJSON(t, fetchedResponse.Body, &fetched)
	if fetched.ID != created.ID || len(fetched.Items) != 2 || fetched.Items[0].ID == "" {
		t.Fatalf("persisted order differs from created order: %+v", fetched)
	}

	listResponse := request(t, http.MethodGet, "/orders", nil)
	defer listResponse.Body.Close()
	if listResponse.StatusCode != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", listResponse.StatusCode)
	}
	var orders []order
	decodeJSON(t, listResponse.Body, &orders)
	if !containsOrder(orders, created.ID) {
		t.Fatalf("created order %s was not returned by list endpoint", created.ID)
	}
}

func TestOrderContractValidation(t *testing.T) {
	tests := []struct {
		name    string
		payload any
	}{
		{name: "missing customer", payload: map[string]any{"items": []map[string]any{{"product_id": "p1", "quantity": 1, "price": 10}}}},
		{name: "empty items", payload: map[string]any{"customer_id": "customer", "items": []any{}}},
		{name: "zero quantity", payload: map[string]any{"customer_id": "customer", "items": []map[string]any{{"product_id": "p1", "quantity": 0, "price": 10}}}},
		{name: "negative price", payload: map[string]any{"customer_id": "customer", "items": []map[string]any{{"product_id": "p1", "quantity": 1, "price": -1}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := request(t, http.MethodPost, "/orders", test.payload)
			defer response.Body.Close()
			if response.StatusCode < 400 || response.StatusCode >= 500 {
				t.Fatalf("expected 4xx, got %d: %s", response.StatusCode, readBody(response.Body))
			}
		})
	}
}

func TestUnknownOrderReturnsNotFound(t *testing.T) {
	response := request(t, http.MethodGet, "/orders/00000000-0000-4000-8000-000000000000", nil)
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", response.StatusCode, readBody(response.Body))
	}
}

func TestMalformedOrderIDReturnsBadRequest(t *testing.T) {
	response := request(t, http.MethodGet, "/orders/does-not-exist", nil)
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.StatusCode, readBody(response.Body))
	}
}

func request(t *testing.T, method, path string, payload any) *http.Response {
	t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(t.Context(), method, baseURL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return response
}

func waitUntilReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		response, err := httpClient.Get(baseURL + "/ready")
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("backend did not become ready at %s within %s", baseURL, timeout)
}

func decodeJSON(t *testing.T, reader io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(reader).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func readBody(reader io.Reader) string {
	value, _ := io.ReadAll(reader)
	return string(value)
}

func containsOrder(orders []order, id string) bool {
	for _, value := range orders {
		if value.ID == id {
			return true
		}
	}
	return false
}
