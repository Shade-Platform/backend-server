package middleware

import (
	"context"
	"net/http"
	"shade_web_server/core/trust"
	"time"
)

const geoIPTimeout = 500 * time.Millisecond

// TrustMiddleware runs the trust‑score check before anything else.
func TrustMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := trust.GetIPFromRequest(r)
		ua := r.UserAgent()

		ctx, cancel := context.WithTimeout(r.Context(), geoIPTimeout)
		defer cancel()
		result := trust.DefaultTrustEngine.CalculateTrustScore(ctx, ip, ua)

		// Ask if failed login attempts should penalize further
		if penalized, penalty := trust.FailedTracker.ShouldPenalize(ip); penalized {
			result.Score += penalty
			result.Reasons = append(result.Reasons, "Too many failed login attempts")
		}

		if result.Score < 30 {
			http.Error(w, "Access denied: low trust score", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
