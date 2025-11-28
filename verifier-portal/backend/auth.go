package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles user authentication
type AuthService struct {
	jwtSecret     []byte
	tokenDuration time.Duration
	users         map[string]*User // In production, use a database
}

// User represents a verifier user
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	LastLogin    time.Time
	IsActive     bool
	APIKey       string
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Credentials represents login credentials
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string) *AuthService {
	if jwtSecret == "" {
		// Generate random secret if not provided
		jwtSecret = generateRandomSecret()
	}

	return &AuthService{
		jwtSecret:     []byte(jwtSecret),
		tokenDuration: 24 * time.Hour,
		users:         make(map[string]*User),
	}
}

// Register creates a new user account
func (a *AuthService) Register(email, password string) (*User, error) {
	// Check if user already exists
	for _, user := range a.users {
		if user.Email == email {
			return nil, errors.New("user already exists")
		}
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:           generateUserID(),
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         "verifier",
		CreatedAt:    time.Now(),
		IsActive:     true,
		APIKey:       generateAPIKey(),
	}

	a.users[user.ID] = user

	return user, nil
}

// Login authenticates a user and returns JWT tokens
func (a *AuthService) Login(credentials Credentials) (*TokenPair, error) {
	// Find user by email
	var user *User
	for _, u := range a.users {
		if u.Email == credentials.Email {
			user = u
			break
		}
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	user.LastLogin = time.Now()

	// Generate tokens
	tokenPair, err := a.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokenPair, nil
}

// ValidateToken validates a JWT token and returns the claims
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new access token from a refresh token
func (a *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := a.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user
	user, exists := a.users[claims.UserID]
	if !exists {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new tokens
	return a.generateTokenPair(user)
}

// GetUser retrieves a user by ID
func (a *AuthService) GetUser(userID string) (*User, error) {
	user, exists := a.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// UpdateUser updates user information
func (a *AuthService) UpdateUser(userID string, updates map[string]interface{}) error {
	user, exists := a.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// Update fields
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}

	if role, ok := updates["role"].(string); ok {
		user.Role = role
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		user.IsActive = isActive
	}

	return nil
}

// ChangePassword changes a user's password
func (a *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	user, exists := a.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// Verify old password
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("incorrect old password")
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	user.PasswordHash = string(newPasswordHash)

	return nil
}

// ResetPassword resets a user's password (admin function)
func (a *AuthService) ResetPassword(userID string) (string, error) {
	user, exists := a.users[userID]
	if !exists {
		return "", errors.New("user not found")
	}

	// Generate temporary password
	tempPassword := generateRandomString(16)

	// Hash temporary password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(passwordHash)

	return tempPassword, nil
}

// ValidateAPIKey validates an API key and returns the associated user
func (a *AuthService) ValidateAPIKey(apiKey string) (*User, error) {
	for _, user := range a.users {
		if user.APIKey == apiKey && user.IsActive {
			return user, nil
		}
	}
	return nil, errors.New("invalid API key")
}

// RegenerateAPIKey generates a new API key for a user
func (a *AuthService) RegenerateAPIKey(userID string) (string, error) {
	user, exists := a.users[userID]
	if !exists {
		return "", errors.New("user not found")
	}

	newAPIKey := generateAPIKey()
	user.APIKey = newAPIKey

	return newAPIKey, nil
}

// Helper functions

func (a *AuthService) generateTokenPair(user *User) (*TokenPair, error) {
	expiresAt := time.Now().Add(a.tokenDuration)

	// Create access token
	accessClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(a.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Create refresh token (longer expiry)
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(a.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func generateRandomSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}
