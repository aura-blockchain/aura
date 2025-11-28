package main

import (
	"errors"
	"sync"
	"time"
)

// VerifierRegistry manages verifier registration and onboarding
type VerifierRegistry struct {
	mu                sync.RWMutex
	verifiers         map[string]*Verifier
	pendingRequests   map[string]*RegistrationRequest
	verificationQueue *Queue
}

// Verifier represents a registered verifier
type Verifier struct {
	ID               string
	UserID           string
	Status           VerifierStatus
	Tier             VerifierTier
	Specializations  []string
	Certifications   []Certification
	TotalVerified    int64
	TotalEarned      int64
	AverageRating    float64
	ReputationScore  int
	JoinedAt         time.Time
	LastActiveAt     time.Time
	Address          string // Blockchain address
	IsActive         bool
	Statistics       VerifierStats
}

// VerifierStatus represents verifier account status
type VerifierStatus string

const (
	StatusPending  VerifierStatus = "pending"
	StatusActive   VerifierStatus = "active"
	StatusSuspended VerifierStatus = "suspended"
	StatusInactive VerifierStatus = "inactive"
)

// VerifierTier represents verifier tier/level
type VerifierTier string

const (
	TierBronze   VerifierTier = "bronze"
	TierSilver   VerifierTier = "silver"
	TierGold     VerifierTier = "gold"
	TierPlatinum VerifierTier = "platinum"
)

// Certification represents a verifier certification
type Certification struct {
	Type      string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Issuer    string
}

// VerifierStats contains verifier statistics
type VerifierStats struct {
	TotalSubmissions    int64
	ApprovedSubmissions int64
	RejectedSubmissions int64
	PendingSubmissions  int64
	AverageProcessTime  time.Duration
	AccuracyRate        float64
}

// RegistrationRequest represents a verifier registration request
type RegistrationRequest struct {
	ID              string
	UserID          string
	Email           string
	FullName        string
	Organization    string
	Specializations []string
	Experience      string
	Documents       []Document
	Status          string
	SubmittedAt     time.Time
	ReviewedAt      *time.Time
	ReviewedBy      string
	Notes           string
}

// Document represents an uploaded document
type Document struct {
	ID          string
	Type        string
	FileName    string
	URL         string
	UploadedAt  time.Time
	VerifiedAt  *time.Time
}

// NewVerifierRegistry creates a new verifier registry
func NewVerifierRegistry() *VerifierRegistry {
	return &VerifierRegistry{
		verifiers:         make(map[string]*Verifier),
		pendingRequests:   make(map[string]*RegistrationRequest),
		verificationQueue: NewQueue(),
	}
}

// SubmitRegistration submits a new verifier registration request
func (vr *VerifierRegistry) SubmitRegistration(req *RegistrationRequest) (string, error) {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	// Check if user already has a pending or active registration
	for _, existing := range vr.pendingRequests {
		if existing.UserID == req.UserID {
			return "", errors.New("registration already pending")
		}
	}

	for _, verifier := range vr.verifiers {
		if verifier.UserID == req.UserID {
			return "", errors.New("already registered as verifier")
		}
	}

	// Generate ID
	req.ID = generateID("reg")
	req.SubmittedAt = time.Now()
	req.Status = "pending_review"

	vr.pendingRequests[req.ID] = req

	return req.ID, nil
}

// ApproveRegistration approves a registration request
func (vr *VerifierRegistry) ApproveRegistration(requestID, reviewerID string) (*Verifier, error) {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	req, exists := vr.pendingRequests[requestID]
	if !exists {
		return nil, errors.New("registration request not found")
	}

	// Create verifier
	verifier := &Verifier{
		ID:              generateID("verifier"),
		UserID:          req.UserID,
		Status:          StatusActive,
		Tier:            TierBronze, // Start at bronze
		Specializations: req.Specializations,
		JoinedAt:        time.Now(),
		LastActiveAt:    time.Now(),
		IsActive:        true,
		ReputationScore: 100, // Starting reputation
		Statistics: VerifierStats{
			AccuracyRate: 100.0,
		},
	}

	vr.verifiers[verifier.ID] = verifier

	// Update request status
	now := time.Now()
	req.Status = "approved"
	req.ReviewedAt = &now
	req.ReviewedBy = reviewerID

	// Remove from pending
	delete(vr.pendingRequests, requestID)

	return verifier, nil
}

// RejectRegistration rejects a registration request
func (vr *VerifierRegistry) RejectRegistration(requestID, reviewerID, reason string) error {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	req, exists := vr.pendingRequests[requestID]
	if !exists {
		return errors.New("registration request not found")
	}

	now := time.Now()
	req.Status = "rejected"
	req.ReviewedAt = &now
	req.ReviewedBy = reviewerID
	req.Notes = reason

	delete(vr.pendingRequests, requestID)

	return nil
}

// GetVerifier retrieves a verifier by ID
func (vr *VerifierRegistry) GetVerifier(verifierID string) (*Verifier, error) {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	verifier, exists := vr.verifiers[verifierID]
	if !exists {
		return nil, errors.New("verifier not found")
	}

	return verifier, nil
}

// GetVerifierByUserID retrieves a verifier by user ID
func (vr *VerifierRegistry) GetVerifierByUserID(userID string) (*Verifier, error) {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	for _, verifier := range vr.verifiers {
		if verifier.UserID == userID {
			return verifier, nil
		}
	}

	return nil, errors.New("verifier not found")
}

// UpdateVerifierStatus updates a verifier's status
func (vr *VerifierRegistry) UpdateVerifierStatus(verifierID string, status VerifierStatus) error {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	verifier, exists := vr.verifiers[verifierID]
	if !exists {
		return errors.New("verifier not found")
	}

	verifier.Status = status

	if status == StatusActive {
		verifier.IsActive = true
	} else {
		verifier.IsActive = false
	}

	return nil
}

// UpdateReputationScore updates a verifier's reputation score
func (vr *VerifierRegistry) UpdateReputationScore(verifierID string, delta int) error {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	verifier, exists := vr.verifiers[verifierID]
	if !exists {
		return errors.New("verifier not found")
	}

	verifier.ReputationScore += delta

	// Ensure score stays within bounds
	if verifier.ReputationScore < 0 {
		verifier.ReputationScore = 0
	}
	if verifier.ReputationScore > 1000 {
		verifier.ReputationScore = 1000
	}

	// Update tier based on reputation
	vr.updateTier(verifier)

	return nil
}

// UpdateStatistics updates verifier statistics
func (vr *VerifierRegistry) UpdateStatistics(verifierID string, verified bool, processingTime time.Duration) error {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	verifier, exists := vr.verifiers[verifierID]
	if !exists {
		return errors.New("verifier not found")
	}

	verifier.Statistics.TotalSubmissions++
	verifier.TotalVerified++

	if verified {
		verifier.Statistics.ApprovedSubmissions++
	} else {
		verifier.Statistics.RejectedSubmissions++
	}

	// Update average processing time
	totalTime := verifier.Statistics.AverageProcessTime * time.Duration(verifier.Statistics.TotalSubmissions-1)
	verifier.Statistics.AverageProcessTime = (totalTime + processingTime) / time.Duration(verifier.Statistics.TotalSubmissions)

	// Update accuracy rate
	if verifier.Statistics.TotalSubmissions > 0 {
		verifier.Statistics.AccuracyRate = (float64(verifier.Statistics.ApprovedSubmissions) / float64(verifier.Statistics.TotalSubmissions)) * 100
	}

	verifier.LastActiveAt = time.Now()

	return nil
}

// AddCertification adds a certification to a verifier
func (vr *VerifierRegistry) AddCertification(verifierID string, cert Certification) error {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	verifier, exists := vr.verifiers[verifierID]
	if !exists {
		return errors.New("verifier not found")
	}

	verifier.Certifications = append(verifier.Certifications, cert)

	return nil
}

// GetPendingRequests returns all pending registration requests
func (vr *VerifierRegistry) GetPendingRequests() []*RegistrationRequest {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	requests := make([]*RegistrationRequest, 0, len(vr.pendingRequests))
	for _, req := range vr.pendingRequests {
		requests = append(requests, req)
	}

	return requests
}

// GetActiveVerifiers returns all active verifiers
func (vr *VerifierRegistry) GetActiveVerifiers() []*Verifier {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	verifiers := make([]*Verifier, 0)
	for _, verifier := range vr.verifiers {
		if verifier.Status == StatusActive {
			verifiers = append(verifiers, verifier)
		}
	}

	return verifiers
}

// GetTopVerifiers returns top verifiers by reputation
func (vr *VerifierRegistry) GetTopVerifiers(limit int) []*Verifier {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	verifiers := make([]*Verifier, 0, len(vr.verifiers))
	for _, v := range vr.verifiers {
		verifiers = append(verifiers, v)
	}

	// Sort by reputation (simple bubble sort)
	for i := 0; i < len(verifiers); i++ {
		for j := i + 1; j < len(verifiers); j++ {
			if verifiers[j].ReputationScore > verifiers[i].ReputationScore {
				verifiers[i], verifiers[j] = verifiers[j], verifiers[i]
			}
		}
	}

	if len(verifiers) > limit {
		verifiers = verifiers[:limit]
	}

	return verifiers
}

// GetStatistics returns registry statistics
func (vr *VerifierRegistry) GetStatistics() map[string]interface{} {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	totalVerifiers := len(vr.verifiers)
	activeVerifiers := 0
	suspendedVerifiers := 0

	tierCounts := make(map[VerifierTier]int)

	for _, v := range vr.verifiers {
		if v.Status == StatusActive {
			activeVerifiers++
		} else if v.Status == StatusSuspended {
			suspendedVerifiers++
		}

		tierCounts[v.Tier]++
	}

	return map[string]interface{}{
		"total_verifiers":     totalVerifiers,
		"active_verifiers":    activeVerifiers,
		"suspended_verifiers": suspendedVerifiers,
		"pending_requests":    len(vr.pendingRequests),
		"tier_distribution":   tierCounts,
	}
}

// Helper functions

func (vr *VerifierRegistry) updateTier(verifier *Verifier) {
	score := verifier.ReputationScore

	if score >= 800 {
		verifier.Tier = TierPlatinum
	} else if score >= 600 {
		verifier.Tier = TierGold
	} else if score >= 400 {
		verifier.Tier = TierSilver
	} else {
		verifier.Tier = TierBronze
	}
}

func generateID(prefix string) string {
	return prefix + "_" + generateRandomString(16)
}
