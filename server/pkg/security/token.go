// Package security provides token management and rotation functionality.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// TokenManager manages token generation and rotation
type TokenManager struct {
	tokens      map[string]*TokenInfo
	mutex       sync.RWMutex
	rotationInterval time.Duration
	expirationTime   time.Duration
}

// TokenInfo represents token information
type TokenInfo struct {
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	LastUsed    time.Time `json:"last_used"`
	AgentID     string    `json:"agent_id,omitempty"`
	Permissions []string  `json:"permissions"`
	IsActive    bool      `json:"is_active"`
}

// NewTokenManager creates a new token manager
func NewTokenManager(rotationInterval, expirationTime time.Duration) *TokenManager {
	tm := &TokenManager{
		tokens:           make(map[string]*TokenInfo),
		rotationInterval: rotationInterval,
		expirationTime:   expirationTime,
	}

	// Start token rotation routine
	go tm.startTokenRotation()

	return tm
}

// GenerateToken generates a new token
func (tm *TokenManager) GenerateToken(agentID string, permissions []string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %v", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	now := time.Now()

	tokenInfo := &TokenInfo{
		Token:       token,
		CreatedAt:   now,
		ExpiresAt:   now.Add(tm.expirationTime),
		LastUsed:    now,
		AgentID:     agentID,
		Permissions: permissions,
		IsActive:    true,
	}

	tm.mutex.Lock()
	tm.tokens[token] = tokenInfo
	tm.mutex.Unlock()

	return token, nil
}

// ValidateToken validates a token and updates last used time
func (tm *TokenManager) ValidateToken(token string) (*TokenInfo, error) {
	tm.mutex.RLock()
	tokenInfo, exists := tm.tokens[token]
	tm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	if !tokenInfo.IsActive {
		return nil, fmt.Errorf("token is inactive")
	}

	if time.Now().After(tokenInfo.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	// Update last used time
	tm.mutex.Lock()
	tokenInfo.LastUsed = time.Now()
	tm.mutex.Unlock()

	return tokenInfo, nil
}

// RevokeToken revokes a token
func (tm *TokenManager) RevokeToken(token string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tokenInfo, exists := tm.tokens[token]; exists {
		tokenInfo.IsActive = false
		return nil
	}

	return fmt.Errorf("token not found")
}

// RotateToken generates a new token for an existing agent
func (tm *TokenManager) RotateToken(oldToken string) (string, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tokenInfo, exists := tm.tokens[oldToken]
	if !exists {
		return "", fmt.Errorf("old token not found")
	}

	// Generate new token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %v", err)
	}

	newToken := base64.URLEncoding.EncodeToString(tokenBytes)
	now := time.Now()

	// Create new token info
	newTokenInfo := &TokenInfo{
		Token:       newToken,
		CreatedAt:   now,
		ExpiresAt:   now.Add(tm.expirationTime),
		LastUsed:    now,
		AgentID:     tokenInfo.AgentID,
		Permissions: tokenInfo.Permissions,
		IsActive:    true,
	}

	// Deactivate old token
	tokenInfo.IsActive = false

	// Add new token
	tm.tokens[newToken] = newTokenInfo

	return newToken, nil
}

// ListTokens returns all tokens (for admin purposes)
func (tm *TokenManager) ListTokens() []*TokenInfo {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tokens := make([]*TokenInfo, 0, len(tm.tokens))
	for _, tokenInfo := range tm.tokens {
		tokens = append(tokens, tokenInfo)
	}

	return tokens
}

// CleanupExpiredTokens removes expired tokens
func (tm *TokenManager) CleanupExpiredTokens() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := time.Now()
	for token, tokenInfo := range tm.tokens {
		if now.After(tokenInfo.ExpiresAt) {
			delete(tm.tokens, token)
		}
	}
}

// startTokenRotation starts the token rotation routine
func (tm *TokenManager) startTokenRotation() {
	ticker := time.NewTicker(tm.rotationInterval)
	defer ticker.Stop()

	for range ticker.C {
		tm.CleanupExpiredTokens()
		// TODO: Implement automatic token rotation for long-lived tokens
	}
}

// GetTokenStats returns token statistics
func (tm *TokenManager) GetTokenStats() map[string]interface{} {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	activeCount := 0
	expiredCount := 0
	now := time.Now()

	for _, tokenInfo := range tm.tokens {
		if tokenInfo.IsActive {
			if now.After(tokenInfo.ExpiresAt) {
				expiredCount++
			} else {
				activeCount++
			}
		}
	}

	return map[string]interface{}{
		"total_tokens":   len(tm.tokens),
		"active_tokens":  activeCount,
		"expired_tokens": expiredCount,
		"rotation_interval": tm.rotationInterval.String(),
		"expiration_time":   tm.expirationTime.String(),
	}
}

