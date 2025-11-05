// File: internal/utils/rate_limiter.go

// This file contains the rate limiting middleware for the feedback endpoint.

package utils

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	tokens   int
}

type RateLimiter struct {
	visitors        map[string]*visitor
	mu              sync.Mutex
	rate            int
	window          time.Duration
	cleanupInterval time.Duration
}

func NewRateLimiter(rate int, windowSeconds int, cleanupIntervalSeconds int) *RateLimiter {
	rl := &RateLimiter{
		visitors:        make(map[string]*visitor),
		rate:            rate,
		window:          time.Duration(windowSeconds) * time.Second,
		cleanupInterval: time.Duration(cleanupIntervalSeconds) * time.Second,
	}
	go rl.cleanupVisitors()
	return rl
}

func (rl *RateLimiter) getVisitor(key string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, exists := rl.visitors[key]
	if !exists {
		v = &visitor{lastSeen: time.Now(), tokens: rl.rate}
		rl.visitors[key] = v
	}
	return v
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.cleanupInterval)
		rl.mu.Lock()
		for k, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.cleanupInterval {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		v := rl.getVisitor(key)
		now := time.Now()
		elapsed := now.Sub(v.lastSeen)
		v.lastSeen = now

		// refill tokens if window has passed
		if elapsed > rl.window {
			v.tokens = rl.rate
		}

		if v.tokens > 0 {
			v.tokens--
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		}
	}
}
