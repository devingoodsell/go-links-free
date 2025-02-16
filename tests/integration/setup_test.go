package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/devingoodsell/go-links-free/internal/api/handlers"
	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/devingoodsell/go-links-free/internal/config"
	"github.com/devingoodsell/go-links-free/internal/db"
	"github.com/devingoodsell/go-links-free/internal/models"
	"github.com/devingoodsell/go-links-free/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var (
	testDB     *db.DB
	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	// Try to load test environment from .env.test
	envFiles := []string{
		"../../.env.test", // Try test env first
		"../../.env",      // Fall back to regular env
	}

	envLoaded := false
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			envLoaded = true
			break
		}
	}

	// If no env files were loaded, use defaults
	if !envLoaded {
		log.Printf("Warning: No environment files found, using default test configuration")
		// Set default test environment variables
		os.Setenv("PORT", "8081")
		os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/golinks_test?sslmode=disable")
		os.Setenv("JWT_SECRET", "test-secret-key")
		os.Setenv("ENABLE_OKTA_SSO", "false")
	}

	// Setup test database
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	testDB, err = db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clear test database
	if err := clearTestDatabase(testDB.DB); err != nil {
		log.Fatalf("Failed to clear test database: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if testDB != nil {
		testDB.Close()
	}

	os.Exit(code)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Initialize auth service for testing
	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), 24*time.Hour)
	userRepo := models.NewUserRepository(testDB)
	authService := auth.NewAuthService(userRepo, jwtManager, false)

	// Initialize link service
	linkRepo := models.NewLinkRepository(testDB)
	linkService := services.NewLinkService(linkRepo)

	// Add routes
	handlers.AddHealthRoutes(router)
	handlers.AddAuthRoutes(router, authService)
	handlers.AddLinkRoutes(router, linkService, authService.AuthMiddleware())

	return router
}

func clearTestDatabase(db *sql.DB) error {
	tables := []string{
		"request_log_aggregates",
		"request_logs",
		"link_stats",
		"links",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %v", table, err)
		}
	}

	return nil
}

func createTestUser(t *testing.T, router *gin.Engine) (userID int64, token string) {
	// Register a test user
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "testpassword123",
	}
	body, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Token string `json:"token"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Get user ID from token
	claims, err := auth.ParseToken(response.Token, os.Getenv("JWT_SECRET"))
	require.NoError(t, err)

	return claims.UserID, response.Token
}

func resetTestDB(t *testing.T) {
	err := clearTestDatabase(testDB.DB)
	require.NoError(t, err, "Failed to clear test database")
}
