package routers

import (
	"context"
	"encoding/json"
	"net/http"
	"shade_web_server/core/trust"
	"shade_web_server/infrastructure/logger"
	"shade_web_server/middleware"
	"time"

	"github.com/gorilla/mux"
)

// InitializeTrustRouter sets up the trust score route.
func InitializeTrustRouter() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/trust/score", middleware.TrustMiddleware(http.HandlerFunc(getTrustScoreHandler))).Methods("GET")
	return r
}

// getTrustScoreHandler evaluates trust score based on IP and User-Agent
func getTrustScoreHandler(w http.ResponseWriter, r *http.Request) {
	ip := trust.GetIPFromRequest(r)
	ua := r.UserAgent()

	// Calculate base trust score
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()
	result := trust.DefaultTrustEngine.CalculateTrustScore(ctx, ip, ua)

	// Apply failed login penalty if needed
	if penalized, penalty := trust.FailedTracker.ShouldPenalize(ip); penalized {
		result.Score += penalty
		result.Reasons = append(result.Reasons, "Too many failed login attempts")
	}

	failedCount := trust.FailedTracker.GetFailureCount(ip)

	// Logging to Kibana
	logger.Log.WithFields(map[string]interface{}{
		"event":           "trust_check",
		"ip":              result.ClientIP,
		"user_agent":      result.UserAgent,
		"score":           result.Score,
		"reasons":         result.Reasons,
		"country":         result.Country,
		"timezone":        result.Timezone,
		"local_hour":      result.LocalHour,
		"failed_attempts": failedCount,
	}).Info("Trust score evaluated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
