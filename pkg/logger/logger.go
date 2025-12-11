// Package logger provides structured logging for the file storage service.
//
// LEARNING NOTES FOR GO BEGINNERS:
// =================================
// This package demonstrates:
// 1. Interfaces in Go
// 2. Methods and method receivers
// 3. The Gin web framework integration
// 4. Structured logging with Zerolog
// 5. Middleware pattern for HTTP logging
//
// WHY STRUCTURED LOGGING?
// Traditional logging: "User John logged in at 10:15"
// Structured logging: {"level":"info", "user":"John", "action":"login", "time":"10:15"}
//
// Structured logs are:
// - Machine-readable (easy to parse and analyze)
// - Searchable (find all logs for user "John")
// - Aggregatable (count login attempts per minute)
// - Better for production systems
package logger

import (
	"fmt"      // For formatting strings
	"os"       // For accessing stdout/stderr
	"time"     // For timestamps

	"github.com/gin-gonic/gin"       // Gin web framework
	"github.com/rs/zerolog"          // Zerolog is a fast, structured logging library
	"github.com/rs/zerolog/log"      // Global logger instance
)

// =============================================================================
// LOGGER INTERFACE AND STRUCT
// =============================================================================

// Logger wraps zerolog.Logger to provide a custom logging interface.
//
// STRUCT WITH EMBEDDED TYPE:
// The Logger struct has one field: logger of type zerolog.Logger
// This is called "composition" - we're building our logger around zerolog's logger
//
// WHY WRAP ZEROLOG?
// 1. We can add custom methods specific to our application
// 2. We can easily switch logging libraries later if needed
// 3. We can add application-specific context to all logs
type Logger struct {
	logger zerolog.Logger  // The underlying zerolog logger
}

// =============================================================================
// CONSTRUCTOR FUNCTION
// =============================================================================

// New creates a new logger instance with the given configuration.
//
// FUNCTION PARAMETERS:
// - serviceName: Name of the service (e.g., "auth-service", "file-service")
// - environment: Environment name (e.g., "development", "production")
// - level: Log level (e.g., "debug", "info", "warn", "error")
//
// CONSTRUCTOR PATTERN:
// In Go, we don't have constructors like in OOP languages.
// Instead, we use functions named "New" or "NewXxx" that return initialized structs.
// This is a widely-used Go convention.
//
// RETURN TYPE:
// Returns a pointer to Logger (*Logger) not a Logger value.
// This is efficient because:
// - We don't copy the struct when passing it around
// - All receivers of this pointer refer to the same logger instance
func New(serviceName, environment, level string) *Logger {
	// Configure the time format for log timestamps
	// RFC3339Nano is a standard format: 2006-01-02T15:04:05.999999999Z07:00
	// RFC3339Nano includes nanosecond precision
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Parse the log level string into a zerolog.Level
	// This converts "debug" -> zerolog.DebugLevel, "info" -> zerolog.InfoLevel, etc.
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		// If the level string is invalid, default to InfoLevel
		// InfoLevel is a good default - shows important info without being too verbose
		logLevel = zerolog.InfoLevel
	}

	// Declare a variable to hold the output writer
	// In Go, var declares a variable without initializing it
	// We'll initialize it below based on the environment
	var output zerolog.ConsoleWriter

	// In development, use pretty-printed console output
	// In production, use JSON output (easier to parse by log aggregation systems)
	if environment == "development" {
		// ConsoleWriter formats logs for human readability
		// Instead of JSON, it prints colored, formatted output to the console
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,              // Write to standard output
			TimeFormat: time.RFC3339,           // Human-readable time format
		}
	} else {
		// In production, we want JSON output
		// JSON logs are easily parsed by log aggregation systems like ELK, Splunk, etc.
		// Note: We use os.Stdout (not the ConsoleWriter) for JSON output
	}

	// Create the logger instance
	// This is the actual construction of the logger
	var zlog zerolog.Logger

	if environment == "development" {
		// Development logger with pretty console output
		zlog = zerolog.New(output).
			Level(logLevel).              // Set the minimum log level
			With().                        // Begin adding context fields
			Timestamp().                   // Add timestamp to every log
			Str("service", serviceName).   // Add service name
			Str("environment", environment). // Add environment
			Logger()                       // Finalize the logger
	} else {
		// Production logger with JSON output
		zlog = zerolog.New(os.Stdout).
			Level(logLevel).
			With().
			Timestamp().
			Str("service", serviceName).
			Str("environment", environment).
			Logger()
	}

	// Return a pointer to our Logger struct
	// The & operator creates a pointer to the struct we're creating
	return &Logger{
		logger: zlog,  // Assign the zerolog logger to our struct
	}
}

// =============================================================================
// LOGGING METHODS
// =============================================================================
// These methods provide an interface for logging at different levels.
//
// METHOD EXPLANATION:
// func (l *Logger) MethodName() returnType
//
// (l *Logger) is the "receiver" - it makes this a method of the Logger type
// l is the variable name we use to access the Logger instance
// *Logger means it's a pointer receiver (can modify the struct if needed)
//
// METHOD CHAINING:
// These methods return *zerolog.Event, which allows method chaining:
// log.Info().Str("user", "john").Msg("user logged in")
// =============================================================================

// Info logs an informational message.
//
// INFO LEVEL:
// Use for general informational messages about application operation.
// Examples: "Server started", "User logged in", "File uploaded"
//
// RETURN VALUE:
// Returns *zerolog.Event which allows method chaining for adding context.
// Example usage:
//     logger.Info().Str("user_id", "123").Msg("User created")
func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Error logs an error message.
//
// ERROR LEVEL:
// Use when something went wrong but the application can continue.
// Examples: "Failed to send email", "Database query timeout", "Invalid input"
//
// Example usage:
//     logger.Error().Err(err).Msg("Failed to process file")
//                   ^^^^
//                   Err() method adds the error to the log
func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Debug logs a debug message.
//
// DEBUG LEVEL:
// Use for detailed information useful during development/debugging.
// Examples: "Request body: {...}", "Calculated value: 42", "Cache miss"
//
// Debug logs are typically disabled in production for performance.
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Warn logs a warning message.
//
// WARN LEVEL:
// Use for potentially harmful situations that aren't errors.
// Examples: "Disk space low", "Deprecated API used", "Slow query detected"
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Fatal logs a fatal message and exits the application.
//
// FATAL LEVEL:
// Use when the application cannot continue and must exit.
// Examples: "Cannot connect to database", "Required config missing"
//
// WARNING: This will call os.Exit(1) and terminate the application!
// Use sparingly, typically only during application startup.
func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// With creates a child logger with additional context fields.
//
// CONTEXT PROPAGATION:
// This is useful for adding context that applies to multiple log statements.
//
// Example usage:
//     userLogger := logger.With().Str("user_id", "123").Logger()
//     userLogger.Info().Msg("User action")  // All logs include user_id
//     userLogger.Debug().Msg("Another action")  // user_id is automatically included
//
// RETURN TYPE:
// Returns zerolog.Context which you call .Logger() on to get a new logger.
func (l *Logger) With() zerolog.Context {
	return l.logger.With()
}

// =============================================================================
// GIN MIDDLEWARE FOR HTTP LOGGING
// =============================================================================

// GinMiddleware returns a Gin middleware function for logging HTTP requests.
//
// MIDDLEWARE EXPLANATION:
// Middleware is code that runs before/after your HTTP handlers.
// Think of it as a chain: Request -> Middleware1 -> Middleware2 -> Handler
//
// Common uses for middleware:
// - Logging requests (this middleware)
// - Authentication (check if user is logged in)
// - Rate limiting (prevent abuse)
// - CORS (cross-origin resource sharing)
// - Error recovery (catch panics)
//
// MIDDLEWARE PATTERN:
// In Gin, middleware is a function that takes a *gin.Context and calls c.Next()
// to pass control to the next handler in the chain.
//
// RETURN TYPE:
// Returns gin.HandlerFunc which is: func(*gin.Context)
// This is the type signature for all Gin handlers and middleware.
func (l *Logger) GinMiddleware() gin.HandlerFunc {
	// Return an anonymous function (closure)
	// This function will be called for every HTTP request
	return func(c *gin.Context) {
		// Record the start time
		// time.Now() returns the current time
		// We'll use this to calculate request duration
		start := time.Now()

		// Get the request path
		// c.Request is the *http.Request object
		// URL.Path is the path portion of the URL (e.g., "/api/users/123")
		path := c.Request.URL.Path

		// Get the raw query string
		// RawQuery is the query parameters (e.g., "?page=1&limit=10")
		query := c.Request.URL.RawQuery

		// Call the next handler in the chain
		// c.Next() is CRITICAL - it passes control to the next middleware/handler
		// Without this, the request handling stops here!
		c.Next()

		// After the handler completes, log the request
		// Everything after c.Next() runs AFTER the request is handled
		// This is called the "response phase" of middleware

		// Calculate request duration
		// time.Since(start) returns how much time passed since 'start'
		// It returns a time.Duration type
		duration := time.Since(start)

		// Build the log entry with request details
		// Method chaining adds fields to the log entry:
		l.logger.Info().
			Str("method", c.Request.Method).          // HTTP method (GET, POST, etc.)
			Str("path", path).                        // URL path
			Str("query", query).                      // Query parameters
			Int("status", c.Writer.Status()).         // HTTP status code (200, 404, etc.)
			Dur("duration_ms", duration).             // Request duration
			Str("client_ip", c.ClientIP()).           // Client IP address
			Str("user_agent", c.Request.UserAgent()). // Browser/client user agent
			Str("request_id", c.GetString("request_id")). // Request ID (from context)
			Int("body_size", c.Writer.Size()).        // Response body size in bytes
			Msg("HTTP request")                       // The main log message
	}
}

// =============================================================================
// USAGE EXAMPLES
// =============================================================================
//
// 1. Creating a logger:
//
//    logger := logger.New("auth-service", "development", "debug")
//
// 2. Simple logging:
//
//    logger.Info().Msg("Server started")
//    logger.Debug().Msg("Processing request")
//    logger.Error().Msg("Failed to connect to database")
//
// 3. Logging with context:
//
//    logger.Info().
//        Str("user_id", "123").
//        Str("action", "login").
//        Msg("User action")
//
//    Output: {"level":"info","service":"auth-service","time":"...","user_id":"123","action":"login","message":"User action"}
//
// 4. Logging errors:
//
//    err := doSomething()
//    if err != nil {
//        logger.Error().
//            Err(err).                    // Add error details
//            Str("operation", "doSomething").
//            Msg("Operation failed")
//    }
//
// 5. Using as Gin middleware:
//
//    router := gin.New()
//    router.Use(logger.GinMiddleware())  // Add logging to all routes
//
//    router.GET("/users", func(c *gin.Context) {
//        c.JSON(200, gin.H{"message": "hello"})
//    })
//
// 6. Creating child loggers with context:
//
//    userLogger := logger.With().
//        Str("user_id", "123").
//        Str("session_id", "abc").
//        Logger()
//
//    userLogger.Info().Msg("User logged in")  // Includes user_id and session_id
//    userLogger.Debug().Msg("Processing...")  // Also includes user_id and session_id
//
// =============================================================================
// LOG LEVELS GUIDE
// =============================================================================
//
// FATAL: Application cannot continue, immediate exit
//    - "Cannot connect to database"
//    - "Required configuration missing"
//    - "Out of memory"
//
// ERROR: Something went wrong, but application continues
//    - "Failed to send email"
//    - "Database query failed"
//    - "File not found"
//
// WARN: Potentially harmful situation, not an error
//    - "Slow query detected (>1s)"
//    - "Using deprecated API"
//    - "Disk space low (10% remaining)"
//
// INFO: General informational messages
//    - "Server started on port 8080"
//    - "User 'john' logged in"
//    - "File uploaded: document.pdf"
//
// DEBUG: Detailed information for debugging
//    - "Request body: {json...}"
//    - "Cache hit for key 'user:123'"
//    - "Calculated result: 42"
//
// =============================================================================
// ZEROLOG PERFORMANCE
// =============================================================================
//
// Zerolog is designed for high performance:
// - Zero allocations (doesn't create garbage for the GC to collect)
// - Lazy evaluation (only processes logs at configured level)
// - Fast JSON encoding (custom, optimized encoder)
//
// Benchmarks show zerolog is ~10x faster than standard logging libraries
// This matters when you log thousands of requests per second!
//
// =============================================================================
