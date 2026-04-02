package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func() *gin.Engine {
		router := gin.New()
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		protected := router.Group("/protected")
		protected.Use(authMiddleware())
		protected.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		return router
	}

	t.Run("health does not require auth", func(t *testing.T) {
		t.Setenv(authTokenEnvVar, "")

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("returns service unavailable when auth is not configured", func(t *testing.T) {
		t.Setenv(authTokenEnvVar, "")

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		assertErrorCode(t, rec, http.StatusServiceUnavailable, "AUTH_NOT_CONFIGURED")
	})

	t.Run("returns unauthorized when header is missing", func(t *testing.T) {
		t.Setenv(authTokenEnvVar, "test-token")

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		assertErrorCode(t, rec, http.StatusUnauthorized, "UNAUTHORIZED")
	})

	t.Run("returns unauthorized when token is invalid", func(t *testing.T) {
		t.Setenv(authTokenEnvVar, "test-token")

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		assertErrorCode(t, rec, http.StatusUnauthorized, "UNAUTHORIZED")
	})

	t.Run("allows request when token is valid", func(t *testing.T) {
		t.Setenv(authTokenEnvVar, "test-token")

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		rec := httptest.NewRecorder()

		newRouter().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}

	if resp.Code != wantCode {
		t.Fatalf("code = %q, want %q", resp.Code, wantCode)
	}
}
