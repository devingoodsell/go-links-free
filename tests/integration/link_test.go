package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type createLinkRequest struct {
	Alias          string     `json:"alias"`
	DestinationURL string     `json:"destination_url"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type linkResponse struct {
	ID             int64      `json:"id"`
	Alias          string     `json:"alias"`
	DestinationURL string     `json:"destination_url"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

func TestLinkOperations(t *testing.T) {
	resetTestDB(t)
	router := setupTestRouter()
	userID, token := createTestUser(t, router)
	_ = userID // We'll use this later for verification if needed

	t.Run("Create Link", func(t *testing.T) {
		link := createLinkRequest{
			Alias:          "google",
			DestinationURL: "https://google.com",
		}
		body, _ := json.Marshal(link)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response linkResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, link.Alias, response.Alias)
		assert.Equal(t, link.DestinationURL, response.DestinationURL)
	})

	t.Run("Get Link", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/go/google", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "https://google.com", w.Header().Get("Location"))
	})

	t.Run("Create Duplicate Link", func(t *testing.T) {
		link := createLinkRequest{
			Alias:          "google",
			DestinationURL: "https://google.com",
		}
		body, _ := json.Marshal(link)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("List Links", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/links", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Links []linkResponse `json:"links"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Links)
	})

	t.Run("Update Link", func(t *testing.T) {
		link := createLinkRequest{
			DestinationURL: "https://www.google.com",
		}
		body, _ := json.Marshal(link)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/links/google", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response linkResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, link.DestinationURL, response.DestinationURL)
	})

	t.Run("Delete Link", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/links/google", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Small delay to ensure deletion is complete
		time.Sleep(100 * time.Millisecond)

		// Verify link is deleted
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/go/google", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestLinkExpiration(t *testing.T) {
	resetTestDB(t)
	router := setupTestRouter()
	_, token := createTestUser(t, router)

	t.Run("Create and Access Expiring Link", func(t *testing.T) {
		expireTime := time.Now().Add(2 * time.Second)
		link := createLinkRequest{
			Alias:          "quick-expire",
			DestinationURL: "https://example.com",
			ExpiresAt:      &expireTime,
		}

		// Create link
		w := httptest.NewRecorder()
		body, _ := json.Marshal(link)
		req, _ := http.NewRequest("POST", "/api/links", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// Access before expiration
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/go/quick-expire", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Access after expiration
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/go/quick-expire", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusGone, w.Code)
	})
}

func TestLinkStats(t *testing.T) {
	resetTestDB(t)
	router := setupTestRouter()
	_, token := createTestUser(t, router)

	t.Run("Track Link Usage Statistics", func(t *testing.T) {
		// Create test link
		link := createLinkRequest{
			Alias:          "stats-test",
			DestinationURL: "https://example.com",
		}
		body, _ := json.Marshal(link)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// Access link multiple times
		for i := 0; i < 5; i++ {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/go/stats-test", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		}

		// Check stats
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/links/stats-test/stats", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var stats struct {
			DailyCount  int `json:"daily_count"`
			WeeklyCount int `json:"weekly_count"`
			TotalCount  int `json:"total_count"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		require.NoError(t, err)
		assert.Equal(t, 5, stats.DailyCount)
		assert.Equal(t, 5, stats.WeeklyCount)
		assert.Equal(t, 5, stats.TotalCount)
	})
}

func TestLinkFiltering(t *testing.T) {
	resetTestDB(t)
	router := setupTestRouter()
	_, token := createTestUser(t, router)

	// Create test links
	testLinks := []createLinkRequest{
		{
			Alias:          "google",
			DestinationURL: "https://google.com",
		},
		{
			Alias:          "example",
			DestinationURL: "https://example.com",
		},
	}

	for _, link := range testLinks {
		body, _ := json.Marshal(link)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", bytes.NewBuffer(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	filters := []struct {
		name   string
		query  string
		expect int
	}{
		{"Active Links", "status=active", 2},
		{"Expired Links", "status=expired", 0},
		{"Search by Domain", "domain=example.com", 1},
		{"Sort by Clicks", "sort=clicks", 2},
	}

	for _, f := range filters {
		t.Run(f.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/links?"+f.query, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Links []linkResponse `json:"links"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Len(t, response.Links, f.expect)
		})
	}
}
