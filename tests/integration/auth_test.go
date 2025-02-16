package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authResponse struct {
	Token string `json:"token"`
}

func TestAuthFlow(t *testing.T) {
	resetTestDB(t)
	router := setupTestRouter()

	// Test data
	testUser := map[string]string{
		"email":    "test@example.com",
		"password": "testpassword123",
	}

	t.Run("Register", func(t *testing.T) {
		body, _ := json.Marshal(testUser)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response authResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Token)
	})

	t.Run("Login", func(t *testing.T) {
		body, _ := json.Marshal(testUser)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response authResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Token)
	})

	t.Run("Register Duplicate Email", func(t *testing.T) {
		body, _ := json.Marshal(testUser)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Login Invalid Password", func(t *testing.T) {
		invalidUser := map[string]string{
			"email":    testUser["email"],
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(invalidUser)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// Helper function to get a valid JWT token for protected endpoint tests
func getTestUserToken(t *testing.T, router *gin.Engine) string {
	testUser := map[string]string{
		"email":    "test@example.com",
		"password": "testpassword123",
	}

	body, _ := json.Marshal(testUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var response authResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	return response.Token
}
