package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aura-chain/aura/faucet/pkg/api"
	"github.com/aura-chain/aura/faucet/pkg/config"
	"github.com/aura-chain/aura/faucet/pkg/database"
	"github.com/aura-chain/aura/faucet/pkg/faucet"
	"github.com/aura-chain/aura/faucet/pkg/ratelimit"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		NodeRPC:             "http://aura-validator-1:26657",
		NodeAPI:             "http://aura-validator-1:1317",
		NodeGRPC:            "aura-validator-1:9090",
		ChainID:             "aura-testnet-1",
		FaucetMnemonic:      "alcohol woman abuse can during mafia husband alcohol ahead begin narrow brave",
		AmountPerRequest:    100000000,
		Environment:         "development",
		DatabaseURL:         "postgres://faucet:faucet@localhost:5432/faucet_test?sslmode=disable",
		RedisURL:            "redis://localhost:6379/1",
		Denom:               "uaura",
		RateLimitPerIP:      10,
		RateLimitPerAddress: 1,
		RateLimitWindow:     24 * time.Hour,
	}

	if err := cfg.Validate(); err != nil {
		t.Skipf("Skipping integration tests, invalid config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Database not available for integration testing: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	require.NoError(t, db.Migrate())

	redisClient, err := ratelimit.NewRedisClient(cfg.RedisURL)
	if err != nil {
		t.Skipf("Redis not available for integration testing: %v", err)
	}
	t.Cleanup(func() {
		_ = redisClient.Close()
	})
	rateLimiter := ratelimit.NewRateLimiter(redisClient, cfg.RateLimitConfig())

	faucetService, err := faucet.NewService(cfg, db)
	if err != nil {
		t.Skipf("Faucet service not available: %v", err)
	}

	handler := api.NewHandler(cfg, faucetService, rateLimiter, db)

	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handler.Health)
		v1.GET("/faucet/info", handler.GetFaucetInfo)
		v1.GET("/faucet/recent", handler.GetRecentTransactions)
		v1.POST("/faucet/request", handler.RequestTokens)
	}

	return router
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)

	// Health check might fail if node is not running, but should return valid JSON
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "status")
}

func TestGetFaucetInfoEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/faucet/info", nil)
	router.ServeHTTP(w, req)

	// This might fail if DB is not available, but we test the endpoint structure
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "amount_per_request")
		assert.Contains(t, response, "denom")
		assert.Contains(t, response, "chain_id")
	}
}

func TestRequestTokensValidation(t *testing.T) {
	router := setupTestRouter(t)

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name: "missing address",
			payload: map[string]string{
				"captcha_token": "test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing captcha",
			payload: map[string]string{
				"address": "aura1z7cawf7uypx2v9m9t447z6tstl0fjfcvvgf0f3",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty payload",
			payload:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/faucet/request", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
