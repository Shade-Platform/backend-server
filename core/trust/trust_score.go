package trust

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TrustResult represents the result of trust evaluation
type TrustResult struct {
	Score     int      `json:"score"`
	Reasons   []string `json:"reasons,omitempty"`
	Country   string   `json:"country,omitempty"`
	Timezone  string   `json:"timezone,omitempty"`
	LocalHour int      `json:"local_hour,omitempty"`
	ClientIP  string   `json:"client_ip"`
	UserAgent string   `json:"user_agent"`
}

// GeoIPInfo holds geographical information from IP
type GeoIPInfo struct {
	Country  string `json:"country_name"`
	Timezone string `json:"timezone"`
}

// GeoIPResolver defines the interface for IP geolocation services
type GeoIPResolver interface {
	Resolve(ctx context.Context, ip string) (GeoIPInfo, int, error)
}

// TrustEngineConfig holds configuration for trust scoring
type TrustEngineConfig struct {
	BadUserAgents        []string
	SuspiciousUserAgents []string
	AbnormalHourStart    int
	AbnormalHourEnd      int
	MaxScore             int
	MinScore             int
	BadUAPenalty         int
	SuspiciousUAPenalty  int
	AbnormalHourPenalty  int
	GeoIPServiceURL      string
}

// TrustEngine handles trust score calculations
type TrustEngine struct {
	config       TrustEngineConfig
	resolver     GeoIPResolver
	uaCheckCache map[string]cachedUAPenalty
	cacheMutex   sync.RWMutex
	cacheTTL     time.Duration
}

type cachedUAPenalty struct {
	penalty int
	expiry  time.Time
}

// IPAPIResolver implements GeoIPResolver using ipapi.co
type IPAPIResolver struct {
	client   *http.Client
	endpoint string
}

// NewDefaultConfig creates a default trust engine configuration
func NewDefaultConfig() TrustEngineConfig {
	return TrustEngineConfig{
		BadUserAgents: []string{
			"sqlmap", "curl", "python-requests", "nmap", "nikto", "wpscan",
		},
		SuspiciousUserAgents: []string{
			"Go-http-client", "Java/", "libwww-perl", "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)",
		},
		AbnormalHourStart:   1,
		AbnormalHourEnd:     5,
		MaxScore:            100,
		MinScore:            0,
		BadUAPenalty:        -100, // Immediate failure
		SuspiciousUAPenalty: -30,  // Suspicious UA penalty
		AbnormalHourPenalty: -40,  // Off-hours penalty
		GeoIPServiceURL:     "https://ipapi.co/%s/json/",
	}
}

// NewTrustEngine creates a new trust engine instance
func NewTrustEngine(config TrustEngineConfig, resolver GeoIPResolver) *TrustEngine {
	if resolver == nil {
		resolver = &IPAPIResolver{
			client: &http.Client{
				Timeout: 2 * time.Second,
				Transport: &http.Transport{
					DisableKeepAlives: true,
					MaxIdleConns:      10,
				},
			},
			endpoint: config.GeoIPServiceURL,
		}
	}

	return &TrustEngine{
		config:       config,
		resolver:     resolver,
		uaCheckCache: make(map[string]cachedUAPenalty),
		cacheTTL:     1 * time.Hour, // Cache entries expire after 1 hour
	}
}

// NewDefaultTrustEngine creates a trust engine with default configuration
func NewDefaultTrustEngine() *TrustEngine {
	config := NewDefaultConfig()
	return NewTrustEngine(config, nil)
}

// Resolve implements GeoIPResolver for IPAPIResolver
func (r *IPAPIResolver) Resolve(ctx context.Context, ip string) (GeoIPInfo, int, error) {
	url := fmt.Sprintf(r.endpoint, ip)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return GeoIPInfo{}, 0, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return GeoIPInfo{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GeoIPInfo{}, resp.StatusCode, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GeoIPInfo{}, resp.StatusCode, err
	}

	var info GeoIPInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return GeoIPInfo{}, resp.StatusCode, err
	}

	return info, resp.StatusCode, nil
}

// CalculateTrustScore computes a trust score for the given IP and User-Agent
func (e *TrustEngine) CalculateTrustScore(ctx context.Context, ip, userAgent string) TrustResult {
	result := TrustResult{
		Score:     e.config.MaxScore,
		ClientIP:  ip,
		UserAgent: userAgent,
	}

	// Check UA cache first
	if penalty := e.getUAPenalty(userAgent); penalty != 0 {
		result.Score += penalty
		result.Reasons = append(result.Reasons, "User-Agent penalty applied")
		if result.Score <= e.config.MinScore {
			result.Score = e.config.MinScore
			return result
		}
	}

	if isPrivateIP(ip) {
		result.Reasons = append(result.Reasons, "Private/reserved IP, skipping GeoIP check")
	} else {
		info, status, err := e.resolver.Resolve(ctx, ip)
		if err != nil {
			result.Reasons = append(result.Reasons, fmt.Sprintf("GeoIP error: %v (status: %d)", err, status))
		} else {
			result.Country = info.Country
			result.Timezone = info.Timezone

			if loc, err := time.LoadLocation(info.Timezone); err == nil {
				result.LocalHour = time.Now().In(loc).Hour()
				if result.LocalHour >= e.config.AbnormalHourStart && result.LocalHour <= e.config.AbnormalHourEnd {
					result.Score += e.config.AbnormalHourPenalty
					result.Reasons = append(result.Reasons,
						fmt.Sprintf("Abnormal access time: %02d:00 local", result.LocalHour))
				}
			} else {
				result.Reasons = append(result.Reasons, "Invalid timezone: "+info.Timezone)
			}
		}
	}

	// Clamp final score
	if result.Score < e.config.MinScore {
		result.Score = e.config.MinScore
	} else if result.Score > e.config.MaxScore {
		result.Score = e.config.MaxScore
	}

	return result
}

// getUAPenalty checks user agent and returns penalty score
func (e *TrustEngine) getUAPenalty(userAgent string) int {
	// Check cache first
	e.cacheMutex.RLock()
	if cached, exists := e.uaCheckCache[userAgent]; exists {
		if time.Now().Before(cached.expiry) {
			e.cacheMutex.RUnlock()
			return cached.penalty
		}
	}
	e.cacheMutex.RUnlock()

	uaLower := strings.ToLower(userAgent)
	penalty := 0

	// Check bad UAs
	for _, bad := range e.config.BadUserAgents {
		if strings.Contains(uaLower, strings.ToLower(bad)) {
			penalty = e.config.BadUAPenalty
			break
		}
	}

	// Check suspicious UAs if no bad match
	if penalty == 0 {
		for _, sus := range e.config.SuspiciousUserAgents {
			if strings.Contains(uaLower, strings.ToLower(sus)) {
				penalty = e.config.SuspiciousUAPenalty
				break
			}
		}
	}

	// Update cache
	e.cacheMutex.Lock()
	e.uaCheckCache[userAgent] = cachedUAPenalty{
		penalty: penalty,
		expiry:  time.Now().Add(e.cacheTTL),
	}
	e.cacheMutex.Unlock()

	return penalty
}

func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true
	}

	privateIPBlocks := []net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},    // loopback
		{IP: net.IPv4(169, 254, 0, 0), Mask: net.CIDRMask(16, 32)}, // link-local
		{IP: net.IPv6loopback, Mask: net.CIDRMask(128, 128)},
		{IP: net.IPv6linklocalallnodes, Mask: net.CIDRMask(128, 128)},
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// GetIPFromRequest extracts client IP from HTTP request
func GetIPFromRequest(r *http.Request) string {
	// Check headers for proxy information
	headers := []string{"X-Forwarded-For", "X-Real-Ip", "CF-Connecting-IP"}
	for _, header := range headers {
		if ip := r.Header.Get(header); ip != "" {
			parts := strings.Split(ip, ",")
			clientIP := strings.TrimSpace(parts[0])
			if net.ParseIP(clientIP) != nil {
				return clientIP
			}
		}
	}

	// Validate RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Fallback for invalid format
		if net.ParseIP(r.RemoteAddr) != nil {
			return r.RemoteAddr
		}
		return "invalid_ip"
	}

	if net.ParseIP(ip) != nil {
		return ip
	}
	return "invalid_ip"
}

var DefaultTrustEngine = NewDefaultTrustEngine()
var FailedTracker = NewFailedLoginTracker(3, 10*time.Minute, -50)
