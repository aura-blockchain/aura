package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// Server represents the verifier portal server
type Server struct {
	router            *mux.Router
	authService       *AuthService
	verifierRegistry  *VerifierRegistry
	queue             *Queue
	rewardsManager    *RewardsManager
	reputationTracker *ReputationTracker
}

// NewServer creates a new server instance
func NewServer() *Server {
	// Initialize services
	jwtSecret := os.Getenv("JWT_SECRET")
	authService := NewAuthService(jwtSecret)

	verifierRegistry := NewVerifierRegistry()

	rewardConfig := RewardConfig{
		BaseReward:        100,
		QualityMultiplier: 1.5,
		SpeedMultiplier:   1.2,
	}
	rewardsManager := NewRewardsManager(rewardConfig)

	reputationTracker := NewReputationTracker()

	server := &Server{
		router:            mux.NewRouter(),
		authService:       authService,
		verifierRegistry:  verifierRegistry,
		queue:             NewQueue(),
		rewardsManager:    rewardsManager,
		reputationTracker: reputationTracker,
	}

	server.setupRoutes()

	return server
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Public routes
	api.HandleFunc("/auth/register", s.handleRegister).Methods("POST")
	api.HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	api.HandleFunc("/auth/refresh", s.handleRefreshToken).Methods("POST")

	// Protected routes (require authentication)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.authMiddleware)

	// Verifier registration
	protected.HandleFunc("/verifier/register", s.handleVerifierRegistration).Methods("POST")
	protected.HandleFunc("/verifier/profile", s.handleGetProfile).Methods("GET")
	protected.HandleFunc("/verifier/stats", s.handleGetStats).Methods("GET")

	// Tasks
	protected.HandleFunc("/tasks", s.handleGetTasks).Methods("GET")
	protected.HandleFunc("/tasks/{id}", s.handleGetTask).Methods("GET")
	protected.HandleFunc("/tasks/{id}/accept", s.handleAcceptTask).Methods("POST")
	protected.HandleFunc("/tasks/{id}/complete", s.handleCompleteTask).Methods("POST")

	// Rewards
	protected.HandleFunc("/rewards", s.handleGetRewards).Methods("GET")
	protected.HandleFunc("/rewards/stats", s.handleGetRewardStats).Methods("GET")

	// Admin routes
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(s.adminMiddleware)
	admin.HandleFunc("/registrations", s.handleGetPendingRegistrations).Methods("GET")
	admin.HandleFunc("/registrations/{id}/approve", s.handleApproveRegistration).Methods("POST")
	admin.HandleFunc("/registrations/{id}/reject", s.handleRejectRegistration).Methods("POST")
	admin.HandleFunc("/verifiers", s.handleGetAllVerifiers).Methods("GET")

	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

// Handler functions

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := s.authService.Register(creds.Email, creds.Password)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	tokens, err := s.authService.Login(creds)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tokens)
}

func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	tokens, err := s.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tokens)
}

func (s *Server) handleVerifierRegistration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	req.UserID = userID

	id, err := s.verifierRegistry.SubmitRegistration(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"registration_id": id,
		"status":          "pending_review",
	})
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	verifier, err := s.verifierRegistry.GetVerifierByUserID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Verifier profile not found")
		return
	}

	respondJSON(w, http.StatusOK, verifier)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	verifier, err := s.verifierRegistry.GetVerifierByUserID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Verifier not found")
		return
	}

	stats := map[string]interface{}{
		"verifier":   verifier,
		"rewards":    s.rewardsManager.GetVerifierTotalEarned(verifier.ID),
		"reputation": s.reputationTracker.GetReputation(verifier.ID),
	}

	respondJSON(w, http.StatusOK, stats)
}

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	verifier, err := s.verifierRegistry.GetVerifierByUserID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Verifier not found")
		return
	}

	tasks := s.queue.GetTasksByVerifier(verifier.ID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	})
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, err := s.queue.GetTask(taskID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Task not found")
		return
	}

	respondJSON(w, http.StatusOK, task)
}

func (s *Server) handleAcceptTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	userID := r.Context().Value("user_id").(string)

	verifier, err := s.verifierRegistry.GetVerifierByUserID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Verifier not found")
		return
	}

	if err := s.queue.Assign(taskID, verifier.ID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.queue.StartTask(taskID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "accepted",
	})
}

func (s *Server) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	userID := r.Context().Value("user_id").(string)

	verifier, err := s.verifierRegistry.GetVerifierByUserID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Verifier not found")
		return
	}

	task, err := s.queue.GetTask(taskID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Task not found")
		return
	}

	if task.AssignedTo != verifier.ID {
		respondError(w, http.StatusForbidden, "Task not assigned to you")
		return
	}

	if err := s.queue.CompleteTask(taskID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Calculate and distribute reward
	completionTime := time.Since(task.CreatedAt)
	quality := 95.0 // Would come from actual verification

	calc := s.rewardsManager.CalculateReward(verifier, task, quality, completionTime)
	dist, err := s.rewardsManager.DistributeReward(
		verifier.ID,
		taskID,
		calc.TotalReward,
		"Task completion",
	)

	if err == nil {
		s.rewardsManager.ProcessReward(dist.ID)
	}

	// Update statistics
	s.verifierRegistry.UpdateStatistics(verifier.ID, true, completionTime)
	s.reputationTracker.UpdateReputation(verifier.ID, 10, "Task completed", taskID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "completed",
		"reward": calc.TotalReward,
	})
}

func (s *Server) handleGetRewards(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	verifier, err := s.verifierRegistry.GetVerifierByUserID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Verifier not found")
		return
	}

	rewards := s.rewardsManager.GetVerifierRewards(verifier.ID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"rewards": rewards,
		"total":   s.rewardsManager.GetVerifierTotalEarned(verifier.ID),
	})
}

func (s *Server) handleGetRewardStats(w http.ResponseWriter, r *http.Request) {
	stats := s.rewardsManager.GetStatistics()
	respondJSON(w, http.StatusOK, stats)
}

func (s *Server) handleGetPendingRegistrations(w http.ResponseWriter, r *http.Request) {
	requests := s.verifierRegistry.GetPendingRequests()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"registrations": requests,
		"count":         len(requests),
	})
}

func (s *Server) handleApproveRegistration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regID := vars["id"]
	reviewerID := r.Context().Value("user_id").(string)

	verifier, err := s.verifierRegistry.ApproveRegistration(regID, reviewerID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, verifier)
}

func (s *Server) handleRejectRegistration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regID := vars["id"]
	reviewerID := r.Context().Value("user_id").(string)

	var req struct {
		Reason string `json:"reason"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	err := s.verifierRegistry.RejectRegistration(regID, reviewerID, req.Reason)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (s *Server) handleGetAllVerifiers(w http.ResponseWriter, r *http.Request) {
	verifiers := s.verifierRegistry.GetActiveVerifiers()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"verifiers": verifiers,
		"count":     len(verifiers),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Middleware

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			respondError(w, http.StatusUnauthorized, "Missing authorization token")
			return
		}

		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := s.authService.ValidateToken(token)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add user ID to context
		ctx := r.Context()
		ctx = contextWithValue(ctx, "user_id", claims.UserID)
		ctx = contextWithValue(ctx, "role", claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value("role").(string)

		if role != "admin" {
			respondError(w, http.StatusForbidden, "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func contextWithValue(ctx interface{}, key, value string) interface{} {
	// In production, use proper context.WithValue
	return ctx
}

func main() {
	server := NewServer()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting verifier portal server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, server.router))
}
