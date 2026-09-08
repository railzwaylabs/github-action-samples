package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cors([]string{"https://app.example.com"}))
	router.GET("/orders", func(c *gin.Context) { c.Status(http.StatusOK) })

	t.Run("allows configured origin", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/orders", nil)
		request.Header.Set("Origin", "https://app.example.com")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
			t.Fatalf("Access-Control-Allow-Origin = %q", got)
		}
	})

	t.Run("handles allowed preflight", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodOptions, "/orders", nil)
		request.Header.Set("Origin", "https://app.example.com")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
	})

	t.Run("rejects unknown preflight origin", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodOptions, "/orders", nil)
		request.Header.Set("Origin", "https://evil.example.com")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
		}
		if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("unexpected Access-Control-Allow-Origin = %q", got)
		}
	})
}
