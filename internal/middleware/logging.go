package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/go-links/internal/auth"
	"github.com/yourusername/go-links/internal/models"
)

type responseWriter struct {
	http.ResponseWriter
	status       int
	wroteHeader  bool
	responseSize int64
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.ResponseWriter.WriteHeader(code)
		rw.wroteHeader = true
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(n)
	return n, err
}

type LoggingMiddleware struct {
	requestLogRepo *models.RequestLogRepository
	logBuffer     []*models.RequestLog
	bufferSize    int
	bufferMutex   sync.Mutex
}

func NewLoggingMiddleware(repo *models.RequestLogRepository) *LoggingMiddleware {
	return &LoggingMiddleware{
		requestLogRepo: repo,
		logBuffer:     make([]*models.RequestLog, 0, 100),
		bufferSize:    100, // Flush after 100 requests
	}
}

func (m *LoggingMiddleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := wrapResponseWriter(w)

		// Generate trace ID
		traceID := uuid.New()

		// Get user ID from context if available
		var userID *int64
		if claims, ok := r.Context().Value("user").(*auth.Claims); ok {
			userID = &claims.UserID
		}

		// Get request size
		var requestSize int64
		if r.ContentLength > 0 {
			requestSize = r.ContentLength
		} else if r.Body != nil {
			// If Content-Length is not set, try to calculate
			body, err := io.ReadAll(r.Body)
			if err == nil {
				requestSize = int64(len(body))
				// Restore body for further processing
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		// Get IP address
		ipAddress := net.ParseIP(getIPAddress(r))

		// Collect request headers
		headers := make(map[string]string)
		for k, v := range r.Header {
			headers[k] = v[0] // Just take the first value for simplicity
		}
		headersJSON, _ := json.Marshal(headers)

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Calculate response time
		duration := time.Since(start)
		responseTime := float64(duration.Microseconds()) / 1000.0

		// Create request log
		log := &models.RequestLog{
			Timestamp:      time.Now(),
			Path:          r.URL.Path,
			Method:        r.Method,
			StatusCode:    wrapped.Status(),
			ResponseTime:  responseTime,
			UserID:        userID,
			IPAddress:     ipAddress,
			UserAgent:     r.UserAgent(),
			Referer:       r.Referer(),
			RequestSize:   requestSize,
			ResponseSize:  wrapped.responseSize,
			Host:          r.Host,
			Protocol:      r.Proto,
			QueryParams:   r.URL.RawQuery,
			RequestHeaders: headersJSON,
			TraceID:       traceID,
		}

		if wrapped.Status() >= 500 {
			errMsg := http.StatusText(wrapped.Status())
			log.ErrorMessage = &errMsg
		}

		// Buffer the log
		m.bufferMutex.Lock()
		m.logBuffer = append(m.logBuffer, log)

		if len(m.logBuffer) >= m.bufferSize {
			go m.flushLogs(m.logBuffer)
			m.logBuffer = make([]*models.RequestLog, 0, m.bufferSize)
		}
		m.bufferMutex.Unlock()
	})
}

func (m *LoggingMiddleware) flushLogs(logs []*models.RequestLog) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.requestLogRepo.CreateBatch(ctx, logs)
	if err != nil {
		// Log error to stderr or error monitoring service
		// For now, just print to stderr
		println("Error flushing request logs:", err.Error())
	}
}

// Helper function to get real IP address
func getIPAddress(r *http.Request) string {
	// Check X-Real-IP header
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Check X-Forwarded-For header
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For might contain multiple IPs, take the first one
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
} 