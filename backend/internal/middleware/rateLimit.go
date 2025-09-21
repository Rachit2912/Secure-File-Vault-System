package middleware

import (
	"backend/internal/config"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Struct for per-user limiter store
type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limiters   = make(map[int]*userLimiter)
	mu         sync.Mutex
	cleanupInt = time.Minute * 5 // cleanup old limiters every 5 minutes 
)

// cleaning inavtive limiters : 
func init() {
	go func() {
		for {
			time.Sleep(cleanupInt)
			mu.Lock()
			for uid, ul := range limiters {
				if time.Since(ul.lastSeen) > cleanupInt {
					delete(limiters, uid)
				}
			}
			mu.Unlock()
		}
	}()
}

// fn. for per-user rate limits :
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // reading value from config file :
        rateLimit := config.AppConfig.ApiRateLimit
        if rateLimit < 0 {
            rateLimit = 2 // default value : 2
        }

        // extract userID from context
        uidVal := r.Context().Value(ContextUserIDKey)
        userID, ok := uidVal.(int)
        if !ok {
            http.Error(w, "unauthorized: userID missing in context", http.StatusUnauthorized)
            return
        }

        // getting or creating limiter for the current user : 
        mu.Lock()
        ul, exists := limiters[userID]
        if !exists {
            ul = &userLimiter{
                limiter:  rate.NewLimiter(rate.Limit(rateLimit), rateLimit),
                lastSeen: time.Now(),
            }
            limiters[userID] = ul
        }
        ul.lastSeen = time.Now()
        mu.Unlock()

        // check allowance :
        if !ul.limiter.Allow() {
            log.Printf("⛔ Rate limit hit for userID=%d", userID)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "rate limit exceeded, try again later",
            })
            return
        }

        // next handler :
        next.ServeHTTP(w, r)
    })
}
