// Package errors provides custom error types for the file storage service.
//
// LEARNING NOTES FOR GO BEGINNERS:
// =================================
// This package demonstrates:
// 1. Error handling in Go (the Go way!)
// 2. Custom error types using structs
// 3. Error wrapping and unwrapping
// 4. HTTP status code mapping
// 5. Interface implementation (the error interface)
//
// ERROR HANDLING IN GO:
// Go doesn't have exceptions (try/catch/throw). Instead:
// - Functions that can fail return an error as their last return value
// - Callers explicitly check if err != nil
// - Errors are values, not exceptions
//
// Example:
//     result, err := doSomething()
//     if err != nil {
//         // Handle the error
//         return err
//     }
//     // Use the result
//
// This explicit approach makes error handling visible and forces you to think
// about what could go wrong. It's more verbose but clearer.
package errors

import (
	"errors"   // Standard library errors package
	"fmt"      // For formatting strings
	"net/http" // For HTTP status codes

	"github.com/gin-gonic/gin" // Gin web framework
)

// =============================================================================
// CUSTOM ERROR TYPE
// =============================================================================

// AppError represents an application-level error with additional context.
//
// STRUCT FIELDS:
// Code:       Machine-readable error code (e.g., "USER_NOT_FOUND")
// Message:    Human-readable error message
// StatusCode: HTTP status code to return (404, 500, etc.)
// Err:        The underlying error (wrapped error)
//
// JSON TAGS:
// The `json:"code"` tags tell the JSON encoder/decoder how to marshal this struct.
// The `json:"-"` tag means "don't include this field in JSON output"
// We exclude StatusCode and Err from JSON because they're internal details.
//
// THE ERROR INTERFACE:
// In Go, any type that has a method:
//     Error() string
// automatically implements the error interface.
// Our AppError implements this below, so it can be used as an error.
type AppError struct {
	Code       string `json:"code"`        // Error code (e.g., "NOT_FOUND")
	Message    string `json:"message"`     // Human-readable message
	StatusCode int    `json:"-"`           // HTTP status code (not in JSON)
	Err        error  `json:"-"`           // Underlying error (not in JSON)
}

// =============================================================================
// ERROR INTERFACE IMPLEMENTATION
// =============================================================================

// Error returns the error message, implementing the error interface.
//
// INTERFACE EXPLANATION:
// Go's error interface is defined in the standard library as:
//
//     type error interface {
//         Error() string
//     }
//
// Any type with an Error() method automatically implements this interface.
// This is called "implicit interface implementation" - you don't declare
// that you're implementing an interface, you just implement its methods.
//
// METHOD RECEIVER:
// (e *AppError) means this method belongs to the AppError type.
// e is the variable name for the AppError instance (like 'this' or 'self')
// * means it's a pointer receiver
//
// RETURN VALUE:
// This method checks if there's a wrapped error (Err).
// If yes: returns "Message: underlying error"
// If no: returns just "Message"
func (e *AppError) Error() string {
	if e.Err != nil {
		// %s is a format verb for strings
		// %v is a format verb for values (works with errors)
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error, supporting error wrapping.
//
// ERROR WRAPPING:
// Go 1.13+ supports error wrapping - storing one error inside another.
// This preserves the error chain so you can see the root cause.
//
// The errors.Is() and errors.As() functions use Unwrap() to check
// the entire error chain.
//
// Example:
//     err1 := errors.New("database error")
//     err2 := Wrap(err1, "failed to save user")
//     errors.Is(err2, err1)  // returns true because err2 wraps err1
func (e *AppError) Unwrap() error {
	return e.Err
}

// =============================================================================
// PREDEFINED ERROR TYPES
// =============================================================================
// These are common errors used throughout the application.
// Defining them here ensures consistent error codes and messages.
//
// VAR EXPLANATION:
// var declares variables
// These are package-level variables (defined outside any function)
// They can be accessed from anywhere as errors.ErrNotFound
//
// POINTER USAGE:
// We use &AppError{...} which creates a pointer to a new AppError struct.
// Using pointers for errors is a Go convention - it's more efficient.
// =============================================================================

var (
	// ErrNotFound indicates a resource was not found (HTTP 404)
	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,  // 404
	}

	// ErrUnauthorized indicates authentication is required (HTTP 401)
	// Use when: User is not logged in, token is missing/invalid
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Unauthorized access",
		StatusCode: http.StatusUnauthorized,  // 401
	}

	// ErrForbidden indicates insufficient permissions (HTTP 403)
	// Use when: User is logged in but doesn't have permission for this action
	// 401 vs 403:
	// - 401: "Who are you?" (authentication)
	// - 403: "I know who you are, but you can't do that" (authorization)
	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "Forbidden",
		StatusCode: http.StatusForbidden,  // 403
	}

	// ErrBadRequest indicates invalid input (HTTP 400)
	// Use when: Request validation fails, malformed JSON, missing required fields
	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "Invalid request",
		StatusCode: http.StatusBadRequest,  // 400
	}

	// ErrConflict indicates a resource conflict (HTTP 409)
	// Use when: User already exists, file already exists, version conflict
	ErrConflict = &AppError{
		Code:       "CONFLICT",
		Message:    "Resource conflict",
		StatusCode: http.StatusConflict,  // 409
	}

	// ErrInternalServer indicates an unexpected server error (HTTP 500)
	// Use when: Database connection failed, unexpected panic recovered, unknown error
	// This is the "something went wrong" error
	ErrInternalServer = &AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Internal server error",
		StatusCode: http.StatusInternalServerError,  // 500
	}

	// ErrServiceUnavailable indicates a temporary service outage (HTTP 503)
	// Use when: Database is down, external API is unreachable, maintenance mode
	// Unlike 500, this suggests the problem is temporary - client should retry
	ErrServiceUnavailable = &AppError{
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "Service temporarily unavailable",
		StatusCode: http.StatusServiceUnavailable,  // 503
	}

	// ErrTooManyRequests indicates rate limiting (HTTP 429)
	// Use when: User exceeded rate limit
	ErrTooManyRequests = &AppError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,  // 429
	}

	// ErrInvalidToken indicates an invalid or expired JWT token
	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "Invalid or expired token",
		StatusCode: http.StatusUnauthorized,  // 401
	}

	// ErrInvalidCredentials indicates wrong username/password
	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid username or password",
		StatusCode: http.StatusUnauthorized,  // 401
	}

	// ErrFileTooLarge indicates file exceeds size limit
	ErrFileTooLarge = &AppError{
		Code:       "FILE_TOO_LARGE",
		Message:    "File exceeds maximum allowed size",
		StatusCode: http.StatusRequestEntityTooLarge,  // 413
	}

	// ErrStorageQuotaExceeded indicates user exceeded storage quota
	ErrStorageQuotaExceeded = &AppError{
		Code:       "STORAGE_QUOTA_EXCEEDED",
		Message:    "Storage quota exceeded",
		StatusCode: http.StatusForbidden,  // 403
	}

	// ErrUnsupportedFileType indicates file type is not supported
	ErrUnsupportedFileType = &AppError{
		Code:       "UNSUPPORTED_FILE_TYPE",
		Message:    "File type is not supported",
		StatusCode: http.StatusBadRequest,  // 400
	}
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// New creates a new AppError with a custom message.
//
// FUNCTION PARAMETERS:
// code:       Error code (e.g., "USER_NOT_FOUND")
// message:    Human-readable message
// statusCode: HTTP status code
//
// USAGE:
//     err := errors.New("USER_NOT_FOUND", "User with ID 123 not found", http.StatusNotFound)
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with an AppError, preserving the error chain.
//
// ERROR WRAPPING EXPLAINED:
// When an error occurs deep in your code, you might want to add context
// as the error bubbles up:
//
// Level 3: db.Query() returns err1: "connection refused"
// Level 2: Wrap(err1, "failed to query database")
// Level 1: Wrap(err2, "failed to get user")
//
// The final error message: "failed to get user: failed to query database: connection refused"
// This shows the entire error chain, making debugging easier.
//
// PARAMETERS:
// err:     The error to wrap (can be any error type)
// message: Additional context to add
//
// USAGE:
//     err := database.Query()
//     if err != nil {
//         return errors.Wrap(err, "failed to query users")
//     }
func Wrap(err error, message string) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,  // Store the original error
	}
}

// WrapWithCode wraps an error with a specific error code and status.
//
// Like Wrap, but allows specifying the error code and HTTP status.
//
// USAGE:
//     err := database.Query()
//     if err != nil {
//         return errors.WrapWithCode(err, "DB_QUERY_FAILED", "Database query failed", 500)
//     }
func WrapWithCode(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Is checks if an error is a specific AppError.
//
// ERROR COMPARISON:
// In Go 1.13+, you can use errors.Is() to check error types in a chain.
// This function makes it easier to check for our AppError types.
//
// USAGE:
//     err := someFunction()
//     if errors.Is(err, errors.ErrNotFound) {
//         // Handle not found error
//     }
//
// RETURN TYPE:
// Returns bool (true/false)
func Is(err error, target *AppError) bool {
	// errors.Is checks if err or any error it wraps matches target
	return errors.Is(err, target)
}

// As extracts an AppError from an error chain.
//
// TYPE ASSERTION WITH ERROR CHAIN:
// If err is or wraps an *AppError, this function extracts it.
//
// USAGE:
//     var appErr *errors.AppError
//     if errors.As(err, &appErr) {
//         // err is an AppError, we can access its fields
//         fmt.Println(appErr.Code)
//         fmt.Println(appErr.StatusCode)
//     }
//
// WHY PASS &appErr?
// We pass a pointer to a pointer (**AppError) so the function can modify
// what appErr points to. This is how Go's errors.As() function works.
func As(err error) (*AppError, bool) {
	var appErr *AppError
	// errors.As attempts to find an *AppError in the error chain
	// If found, it sets appErr to point to it and returns true
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// =============================================================================
// GIN ERROR HANDLER MIDDLEWARE
// =============================================================================

// ErrorHandler is a Gin middleware that handles errors returned by handlers.
//
// HOW IT WORKS:
// 1. Request comes in
// 2. Handler runs (may add errors to c.Errors)
// 3. ErrorHandler middleware runs after the handler
// 4. Checks if any errors occurred
// 5. Converts errors to JSON responses with appropriate HTTP status codes
//
// GIN ERROR HANDLING:
// Gin has a c.Errors slice that collects errors during request processing.
// Handlers can add errors using: c.Error(err)
// This middleware processes those errors and sends appropriate responses.
//
// USAGE:
//     router := gin.New()
//     router.Use(errors.ErrorHandler())  // Add as global middleware
//
// RETURN TYPE:
// Returns gin.HandlerFunc which is the type for all Gin middleware/handlers
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// c.Next() calls the next handler in the chain
		// After it returns, we check if any errors occurred
		c.Next()

		// Check if there are any errors
		// len(c.Errors) returns the number of errors
		// In Go, len() works on slices, arrays, maps, strings, and channels
		if len(c.Errors) == 0 {
			// No errors, nothing to do
			return
		}

		// Get the last error (most recent)
		// c.Errors.Last() returns the last error added
		// .Err extracts the actual error value
		err := c.Errors.Last().Err

		// Try to convert it to our AppError type
		// This uses our As() function defined above
		appErr, ok := As(err)
		if ok {
			// It's an AppError, respond with its details
			// c.JSON sends a JSON response
			c.JSON(appErr.StatusCode, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}

		// Not an AppError, treat as internal server error
		// This handles standard Go errors (from standard library, etc.)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		})
	}
}

// AbortWithError is a helper that aborts the request with an error response.
//
// ABORT EXPLAINED:
// In Gin, c.Abort() stops the middleware chain.
// No subsequent handlers/middleware will run.
// Use when you want to immediately return an error response.
//
// DIFFERENCE FROM c.Error():
// - c.Error() adds an error but continues processing
// - AbortWithError() adds an error AND stops processing
//
// USAGE:
//     if user == nil {
//         errors.AbortWithError(c, errors.ErrNotFound)
//         return
//     }
//
// PARAMETERS:
// c:   Gin context (the current request context)
// err: The AppError to return
func AbortWithError(c *gin.Context, err *AppError) {
	// Abort the request and set the status code
	// This prevents subsequent handlers from running
	c.Abort()

	// Send JSON error response
	c.JSON(err.StatusCode, gin.H{
		"code":    err.Code,
		"message": err.Message,
	})
}

// =============================================================================
// USAGE EXAMPLES
// =============================================================================
//
// 1. Using predefined errors:
//
//    func GetUser(c *gin.Context) {
//        user, err := database.FindUser(id)
//        if err != nil {
//            errors.AbortWithError(c, errors.ErrNotFound)
//            return
//        }
//        c.JSON(200, user)
//    }
//
// 2. Creating custom errors:
//
//    if age < 18 {
//        err := errors.New("AGE_TOO_LOW", "Must be 18 or older", http.StatusBadRequest)
//        errors.AbortWithError(c, err)
//        return
//    }
//
// 3. Wrapping errors:
//
//    result, err := database.Query("SELECT ...")
//    if err != nil {
//        return errors.Wrap(err, "failed to query users")
//    }
//
// 4. Checking error types:
//
//    err := someFunction()
//    if errors.Is(err, errors.ErrNotFound) {
//        // Handle not found specifically
//    } else if errors.Is(err, errors.ErrUnauthorized) {
//        // Handle unauthorized
//    } else {
//        // Handle other errors
//    }
//
// 5. Using as middleware:
//
//    router := gin.New()
//    router.Use(errors.ErrorHandler())
//
//    router.GET("/users/:id", func(c *gin.Context) {
//        user, err := getUser(c.Param("id"))
//        if err != nil {
//            c.Error(err)  // Add error to context
//            return        // ErrorHandler middleware will handle it
//        }
//        c.JSON(200, user)
//    })
//
// =============================================================================
// HTTP STATUS CODE REFERENCE
// =============================================================================
//
// 2xx Success:
// - 200 OK: Request succeeded
// - 201 Created: Resource created successfully
// - 204 No Content: Success but no response body
//
// 4xx Client Errors (client did something wrong):
// - 400 Bad Request: Invalid input, malformed request
// - 401 Unauthorized: Authentication required
// - 403 Forbidden: Authenticated but lacks permission
// - 404 Not Found: Resource doesn't exist
// - 409 Conflict: Resource already exists, version conflict
// - 413 Payload Too Large: File/request too large
// - 429 Too Many Requests: Rate limit exceeded
//
// 5xx Server Errors (server did something wrong):
// - 500 Internal Server Error: Unexpected server error
// - 503 Service Unavailable: Temporary outage, try again later
//
// =============================================================================
