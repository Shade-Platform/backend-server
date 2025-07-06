package middleware

import (
	"context"
	"net/http"
	"shade_web_server/core/trust"
	"time"
)

// const geoIPTimeout = 500 * time.Millisecond

// TrustMiddleware runs the trust‑score check before anything else.
func TrustMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := trust.GetIPFromRequest(r)
		ua := r.UserAgent()

		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		result := trust.DefaultTrustEngine.CalculateTrustScore(ctx, ip, ua)

		if penalized, _ := trust.FailedTracker.ShouldPenalize(ip); penalized {
			result.Score = 0
			result.Reasons = append(result.Reasons, "Too many failed login attempts")
		}

		if result.Score < 30 {
			http.Error(w, "Access denied: low trust score", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
