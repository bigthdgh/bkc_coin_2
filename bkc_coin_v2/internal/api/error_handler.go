package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// ErrorCode represents different error types
type ErrorCode string

const (
	ErrCodeInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeRateLimit       ErrorCode = "RATE_LIMIT"
	ErrCodeInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeValidationError ErrorCode = "VALIDATION_ERROR"
	ErrCodeDuplicateEntry   ErrorCode = "DUPLICATE_ENTRY"
	ErrCodeInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"
	ErrCodeEnergyDepleted   ErrorCode = "ENERGY_DEPLETED"
	ErrCodeDailyLimit      ErrorCode = "DAILY_LIMIT"
	ErrCodeMaintenance     ErrorCode = "MAINTENANCE"
)

// APIError represents a structured API error
type APIError struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ErrorResponse represents the complete error response
type ErrorResponse struct {
	Error   *APIError `json:"error"`
	Success bool      `json:"success"`
}

// ErrorHandler handles HTTP errors with proper formatting
type ErrorHandler struct {
	logger *log.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *log.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError handles an error and writes appropriate response
func (eh *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := middleware.GetReqID(r.Context())
	
	// Determine error type and status code
	apiErr, status := eh.classifyError(err)
	apiErr.RequestID = requestID

	// Log error
	eh.logError(r, apiErr, status)

	// Write response
	eh.writeErrorResponse(w, apiErr, status)
}

// classifyError determines the error type and HTTP status code
func (eh *ErrorHandler) classifyError(err error) (*APIError, int) {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr, eh.getStatusCodeForError(apiErr.Code)
	}

	// Handle common error types
	errMsg := err.Error()
	
	switch {
	case strings.Contains(errMsg, "invalid"):
		return &APIError{
			Code:      ErrCodeInvalidRequest,
			Message:   "Invalid request format",
			Timestamp: time.Now(),
		}, http.StatusBadRequest
		
	case strings.Contains(errMsg, "unauthorized"):
		return &APIError{
			Code:      ErrCodeUnauthorized,
			Message:   "Authentication required",
			Timestamp: time.Now(),
		}, http.StatusUnauthorized
		
	case strings.Contains(errMsg, "forbidden"):
		return &APIError{
			Code:      ErrCodeForbidden,
			Message:   "Access denied",
			Timestamp: time.Now(),
		}, http.StatusForbidden
		
	case strings.Contains(errMsg, "not found"):
		return &APIError{
			Code:      ErrCodeNotFound,
			Message:   "Resource not found",
			Timestamp: time.Now(),
		}, http.StatusNotFound
		
	case strings.Contains(errMsg, "rate limit"):
		return &APIError{
			Code:      ErrCodeRateLimit,
			Message:   "Too many requests",
			Details: map[string]interface{}{
				"retry_after": 60,
			},
			Timestamp: time.Now(),
		}, http.StatusTooManyRequests
		
	case strings.Contains(errMsg, "validation"):
		return &APIError{
			Code:      ErrCodeValidationError,
			Message:   "Validation failed",
			Timestamp: time.Now(),
		}, http.StatusBadRequest
		
	case strings.Contains(errMsg, "duplicate"):
		return &APIError{
			Code:      ErrCodeDuplicateEntry,
			Message:   "Resource already exists",
			Timestamp: time.Now(),
		}, http.StatusConflict
		
	case strings.Contains(errMsg, "insufficient"):
		return &APIError{
			Code:      ErrCodeInsufficientFunds,
			Message:   "Insufficient funds",
			Timestamp: time.Now(),
		}, http.StatusPaymentRequired
		
	case strings.Contains(errMsg, "energy"):
		return &APIError{
			Code:      ErrCodeEnergyDepleted,
			Message:   "Energy depleted",
			Details: map[string]interface{}{
				"recovery_time": "1 hour",
			},
			Timestamp: time.Now(),
		}, http.StatusTooManyRequests
		
	case strings.Contains(errMsg, "daily limit"):
		return &APIError{
			Code:      ErrCodeDailyLimit,
			Message:   "Daily limit reached",
			Details: map[string]interface{}{
				"reset_time": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
			Timestamp: time.Now(),
		}, http.StatusTooManyRequests
		
	case strings.Contains(errMsg, "maintenance"):
		return &APIError{
			Code:      ErrCodeMaintenance,
			Message:   "Service under maintenance",
			Details: map[string]interface{}{
				"estimated_downtime": "30 minutes",
			},
			Timestamp: time.Now(),
		}, http.StatusServiceUnavailable
		
	default:
		return &APIError{
			Code:      ErrCodeInternalError,
			Message:   "Internal server error",
			Timestamp: time.Now(),
		}, http.StatusInternalServerError
	}
}

// getStatusCodeForError returns HTTP status code for error code
func (eh *ErrorHandler) getStatusCodeForError(code ErrorCode) int {
	switch code {
	case ErrCodeInvalidRequest, ErrCodeValidationError:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeDuplicateEntry:
		return http.StatusConflict
	case ErrCodeRateLimit, ErrCodeEnergyDepleted, ErrCodeDailyLimit:
		return http.StatusTooManyRequests
	case ErrCodeInsufficientFunds:
		return http.StatusPaymentRequired
	case ErrCodeServiceUnavailable, ErrCodeMaintenance:
		return http.StatusServiceUnavailable
	case ErrCodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// logError logs the error with context
func (eh *ErrorHandler) logError(r *http.Request, apiErr *APIError, status int) {
	// Don't log client errors (4xx) unless they're validation errors
	if status >= 400 && status < 500 && apiErr.Code != ErrCodeValidationError {
		return
	}

	// Get stack trace for server errors
	var stackTrace string
	if status >= 500 {
		stackTrace = getStackTrace()
	}

	logEntry := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"method":     r.Method,
		"path":       r.URL.Path,
		"status":     status,
		"error_code": apiErr.Code,
		"message":    apiErr.Message,
		"user_agent": r.Header.Get("User-Agent"),
		"ip":        getClientIP(r),
		"request_id": apiErr.RequestID,
	}

	if stackTrace != "" {
		logEntry["stack_trace"] = stackTrace
	}

	if apiErr.Details != nil {
		logEntry["details"] = apiErr.Details
	}

	// Log as JSON for structured logging
	logJSON, _ := json.Marshal(logEntry)
	eh.logger.Printf("ERROR: %s", string(logJSON))
}

// writeErrorResponse writes the error response
func (eh *ErrorHandler) writeErrorResponse(w http.ResponseWriter, apiErr *APIError, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Error:   apiErr,
		Success: false,
	}

	json.NewEncoder(w).Encode(response)
}

// RecoveryMiddleware handles panics and converts them to errors
func (eh *ErrorHandler) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Create error from panic
				apiErr := &APIError{
					Code:      ErrCodeInternalError,
					Message:   "Internal server error",
					Details: map[string]interface{}{
						"panic": fmt.Sprintf("%v", err),
					},
					Timestamp: time.Now(),
				}

				// Log panic
				eh.logger.Printf("PANIC: %v\n%s", err, getStackTrace())

				// Write error response
				eh.writeErrorResponse(w, apiErr, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ValidationMiddleware validates common request parameters
func (eh *ErrorHandler) ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate content type for POST/PUT requests
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				apiErr := &APIError{
					Code:      ErrCodeValidationError,
					Message:   "Content-Type must be application/json",
					Timestamp: time.Now(),
				}
				eh.writeErrorResponse(w, apiErr, http.StatusBadRequest)
				return
			}
		}

		// Validate request size
		if r.ContentLength > 10*1024*1024 { // 10MB limit
			apiErr := &APIError{
				Code:      ErrCodeValidationError,
				Message:   "Request too large",
				Details: map[string]interface{}{
					"max_size": "10MB",
				},
				Timestamp: time.Now(),
			}
			eh.writeErrorResponse(w, apiErr, http.StatusRequestEntityTooLarge)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware implements basic rate limiting
func (eh *ErrorHandler) RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This is a simplified rate limiter
			// In production, use a proper rate limiting library like golang.org/x/time/rate
			
			clientIP := getClientIP(r)
			
			// For demo purposes, we'll use a simple in-memory counter
			// In production, use Redis or similar for distributed rate limiting
			
			// TODO: Implement proper rate limiting
			_ = clientIP
			_ = requestsPerMinute
			
			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func getStackTrace() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// Error creation helpers

func NewValidationError(message string, details map[string]interface{}) *APIError {
	return &APIError{
		Code:      ErrCodeValidationError,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

func NewInvalidRequestError(message string) *APIError {
	return &APIError{
		Code:      ErrCodeInvalidRequest,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		Code:      ErrCodeUnauthorized,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewForbiddenError(message string) *APIError {
	return &APIError{
		Code:      ErrCodeForbidden,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewNotFoundError(message string) *APIError {
	return &APIError{
		Code:      ErrCodeNotFound,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewRateLimitError(retryAfter int) *APIError {
	return &APIError{
		Code:      ErrCodeRateLimit,
		Message:   "Too many requests",
		Details: map[string]interface{}{
			"retry_after": retryAfter,
		},
		Timestamp: time.Now(),
	}
}

func NewInsufficientFundsError(message string) *APIError {
	return &APIError{
		Code:      ErrCodeInsufficientFunds,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewEnergyDepletedError(recoveryTime string) *APIError {
	return &APIError{
		Code:      ErrCodeEnergyDepleted,
		Message:   "Energy depleted",
		Details: map[string]interface{}{
			"recovery_time": recoveryTime,
		},
		Timestamp: time.Now(),
	}
}

func NewDailyLimitError(resetTime time.Time) *APIError {
	return &APIError{
		Code:      ErrCodeDailyLimit,
		Message:   "Daily limit reached",
		Details: map[string]interface{}{
			"reset_time": resetTime.Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	}
}

func NewInternalError(message string) *APIError {
	return &APIError{
		Code:      ErrCodeInternalError,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func NewServiceUnavailableError(message string) *APIError {
	return &APIError{
		Code:      ErrCodeServiceUnavailable,
		Message:   message,
		Timestamp: time.Now(),
	}
}
