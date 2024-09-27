package middlewares

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bugrakocabay/api-gateway/internal/config"
)

// LocalThrottler is responsible for rate limiting based on IP and endpoint.
type LocalThrottler struct {
	routeLimits       map[routeKey]int64
	userRequestCounts sync.Map
}

type routeKey struct {
	method string
	path   string
}

// NewLocalThrottler initializes the LocalThrottler with the provided routes.
func NewLocalThrottler(routes []*config.Route) *LocalThrottler {
	lt := &LocalThrottler{
		routeLimits:       make(map[routeKey]int64),
		userRequestCounts: sync.Map{},
	}
	for _, route := range routes {
		key := routeKey{method: route.Method, path: route.Path}
		lt.routeLimits[key] = route.Limit
	}
	go lt.periodicReset()

	return lt
}

// LocalThrottlerMiddleware applies rate limiting to incoming requests by keeping
// the request counts on memory.
func (lt *LocalThrottler) LocalThrottlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		key := routeKey{method: r.Method, path: r.URL.Path}

		limit, ok := lt.routeLimits[key]
		if !ok || limit == 0 {
			next.ServeHTTP(w, r)
			return
		}

		count := lt.getUserRequestCount(key, ip)

		if count >= limit {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Rate limit exceeded",
			})
			return
		}
		lt.incrementUserRequestCount(key, ip)

		next.ServeHTTP(w, r)
	})
}

type userKey struct {
	endpoint routeKey
	ip       string
}

func (lt *LocalThrottler) incrementUserRequestCount(endpoint routeKey, ip string) {
	key := userKey{endpoint: endpoint, ip: ip}
	val, _ := lt.userRequestCounts.LoadOrStore(key, new(int64))
	countPtr := val.(*int64)
	atomic.AddInt64(countPtr, 1)
}

func (lt *LocalThrottler) getUserRequestCount(endpoint routeKey, ip string) int64 {
	key := userKey{endpoint: endpoint, ip: ip}
	if val, ok := lt.userRequestCounts.Load(key); ok {
		countPtr := val.(*int64)
		return atomic.LoadInt64(countPtr)
	}
	return 0
}

func (lt *LocalThrottler) periodicReset() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		lt.resetCounts()
	}
}

func (lt *LocalThrottler) resetCounts() {
	lt.userRequestCounts.Range(func(key, value interface{}) bool {
		lt.userRequestCounts.Delete(key)
		return true
	})
}

func getClientIP(r *http.Request) string {
	headers := []string{"X-Forwarded-For", "X-Real-IP"}
	for _, h := range headers {
		ips := r.Header.Get(h)
		if ips != "" {
			ip := strings.Split(ips, ",")[0]
			return strings.TrimSpace(ip)
		}
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
