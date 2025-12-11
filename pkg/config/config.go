// Package config provides configuration management for the file storage service.
//
// LEARNING NOTES FOR GO BEGINNERS:
// =================================
// This package demonstrates several important Go concepts:
// 1. Package declaration and imports
// 2. Struct types and field tags
// 3. Functions and error handling
// 4. The Viper library for configuration management
//
// In Go, configuration is typically handled through:
// - Environment variables (for deployment flexibility)
// - Configuration files (for local development)
// - Default values (for sensible fallbacks)
package config

import (
	"fmt" // fmt package for formatting and printing
	"time" // time package for duration types

	"github.com/spf13/viper" // Viper is a popular configuration library for Go
)

// =============================================================================
// CONFIGURATION STRUCT DEFINITIONS
// =============================================================================
// In Go, a struct is a composite data type that groups together related fields.
// Think of it as a blueprint for creating objects with specific properties.
//
// The `mapstructure` tags tell Viper how to map environment variables to fields.
// For example, if you have ENV_VAR=value, the tag `mapstructure:"ENV_VAR"`
// tells Viper to populate that field with the environment variable's value.
// =============================================================================

// Config is the main configuration struct that holds all application settings.
// It's organized into logical sections for easier management.
//
// STRUCT EXPLANATION:
// In Go, we define a struct like this:
//     type StructName struct {
//         FieldName FieldType
//     }
//
// This Config struct contains nested structs (Server, Database, etc.)
// This is called composition - building complex types from simpler ones.
type Config struct {
	Server        ServerConfig        // HTTP server configuration
	Database      DatabaseConfig      // MongoDB database configuration
	S3            S3Config            // AWS S3/MinIO storage configuration
	RabbitMQ      RabbitMQConfig      // RabbitMQ message queue configuration
	Redis         RedisConfig         // Redis cache configuration
	JWT           JWTConfig           // JWT authentication configuration
	RateLimit     RateLimitConfig     // Rate limiting configuration
	Observability ObservabilityConfig // Logging, metrics, tracing configuration
	Worker        WorkerConfig        // Background worker configuration
	Email         EmailConfig         // Email notification configuration
	Security      SecurityConfig      // Security settings
	Services      ServicesConfig      // URLs for inter-service communication
}

// ServerConfig holds HTTP server settings.
//
// FIELD TAG EXPLANATION:
// The `mapstructure:"environment"` tag means this field gets its value from
// an environment variable or config file key named "environment".
//
// Default values are set in the setDefaults() function below.
type ServerConfig struct {
	Name            string        `mapstructure:"service_name"`     // Service name for logging/metrics
	Environment     string        `mapstructure:"environment"`      // dev, staging, prod
	Port            int           `mapstructure:"http_port"`        // HTTP port to listen on
	Version         string        `mapstructure:"version"`          // Application version
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"` // Graceful shutdown timeout
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`     // Max time to read request
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`    // Max time to write response
}

// DatabaseConfig holds MongoDB connection settings.
//
// CONNECTION POOLING EXPLANATION:
// Connection pooling reuses database connections instead of creating new ones
// for each request. This significantly improves performance because:
// - Creating a new connection is expensive (network handshake, authentication)
// - Pooled connections are ready to use immediately
// - You can control how many connections are allowed (prevents overwhelming DB)
type DatabaseConfig struct {
	URI             string        `mapstructure:"mongo_uri"`                // MongoDB connection string
	Database        string        `mapstructure:"mongo_database"`           // Database name
	MaxPoolSize     uint64        `mapstructure:"mongo_max_pool_size"`      // Max connections in pool
	MinPoolSize     uint64        `mapstructure:"mongo_min_pool_size"`      // Min connections in pool
	MaxConnIdleTime time.Duration `mapstructure:"mongo_max_conn_idle_time"` // Max time connection can be idle
	ConnectTimeout  time.Duration `mapstructure:"mongo_connect_timeout"`    // Connection timeout
}

// S3Config holds AWS S3 or MinIO configuration.
//
// PRE-SIGNED URL EXPLANATION:
// Pre-signed URLs allow clients to upload/download files directly to/from S3
// without going through our server. This is much more efficient because:
// - Reduces load on our servers (files don't pass through them)
// - Faster for users (direct connection to S3)
// - Still secure (URLs expire after PresignExpiry time)
type S3Config struct {
	Endpoint        string        `mapstructure:"s3_endpoint"`        // S3 endpoint URL (empty for AWS)
	Region          string        `mapstructure:"aws_region"`         // AWS region
	Bucket          string        `mapstructure:"s3_bucket"`          // Bucket name
	AccessKeyID     string        `mapstructure:"aws_access_key_id"`  // AWS access key
	SecretAccessKey string        `mapstructure:"aws_secret_access_key"` // AWS secret key
	UseSSL          bool          `mapstructure:"s3_use_ssl"`         // Use HTTPS?
	PresignExpiry   time.Duration `mapstructure:"presign_url_expiry"` // Pre-signed URL expiry time
	ChunkSize       int64         `mapstructure:"chunk_size"`         // Chunk size for multipart upload
	MaxFileSize     int64         `mapstructure:"max_file_size"`      // Maximum allowed file size
}

// RabbitMQConfig holds RabbitMQ message broker settings.
//
// MESSAGE QUEUE EXPLANATION:
// RabbitMQ is a message broker that enables asynchronous communication between services.
// Instead of Service A calling Service B directly (synchronous), Service A publishes
// a message to a queue, and Service B processes it when ready (asynchronous).
//
// Benefits:
// - Services are decoupled (don't need to know about each other)
// - Handles traffic spikes (messages queue up if service is busy)
// - Reliability (messages aren't lost if a service is down)
type RabbitMQConfig struct {
	URL            string        `mapstructure:"rabbitmq_url"`             // Connection URL
	Exchange       string        `mapstructure:"rabbitmq_exchange"`        // Exchange name
	PrefetchCount  int           `mapstructure:"rabbitmq_prefetch_count"`  // How many messages to fetch at once
	ReconnectDelay time.Duration `mapstructure:"rabbitmq_reconnect_delay"` // Delay before reconnecting
}

// RedisConfig holds Redis cache configuration.
//
// REDIS EXPLANATION:
// Redis is an in-memory data store (data lives in RAM, not on disk).
// This makes it extremely fast - perfect for:
// - Caching frequently accessed data
// - Session storage
// - Rate limiting (counting requests)
// - Real-time analytics
type RedisConfig struct {
	Host       string `mapstructure:"redis_host"`        // Redis server hostname
	Port       int    `mapstructure:"redis_port"`        // Redis server port
	Password   string `mapstructure:"redis_password"`    // Redis password
	DB         int    `mapstructure:"redis_db"`          // Redis database number (0-15)
	MaxRetries int    `mapstructure:"redis_max_retries"` // Max retry attempts
	PoolSize   int    `mapstructure:"redis_pool_size"`   // Connection pool size
}

// JWTConfig holds JWT authentication settings.
//
// JWT (JSON WEB TOKEN) EXPLANATION:
// JWT is a way to securely transmit information between parties as a JSON object.
// In our case, it's used for authentication:
//
// 1. User logs in with username/password
// 2. Server verifies credentials and generates a JWT token
// 3. Token contains user information (ID, email, roles) and is cryptographically signed
// 4. Client includes token in subsequent requests
// 5. Server verifies token signature and extracts user information
//
// Benefits:
// - Stateless (server doesn't need to store session data)
// - Scalable (any server can verify the token)
// - Secure (tampering is detected via signature verification)
type JWTConfig struct {
	Secret        string        `mapstructure:"jwt_secret"`         // Secret key for signing tokens
	Expiry        time.Duration `mapstructure:"jwt_expiry"`         // Access token expiry
	RefreshExpiry time.Duration `mapstructure:"jwt_refresh_expiry"` // Refresh token expiry
	Issuer        string        `mapstructure:"jwt_issuer"`         // Token issuer identifier
}

// RateLimitConfig holds rate limiting settings.
//
// RATE LIMITING EXPLANATION:
// Rate limiting restricts how many requests a user can make in a given time period.
// This prevents:
// - API abuse (someone making millions of requests)
// - DDoS attacks (distributed denial of service)
// - Accidental infinite loops in client code
//
// TOKEN BUCKET ALGORITHM:
// Imagine a bucket that holds tokens:
// - Bucket has a max capacity (BurstSize)
// - Tokens are added at a steady rate (RequestsPerMin)
// - Each request consumes one token
// - If bucket is empty, request is rejected
// - Allows bursts of traffic as long as average rate is respected
type RateLimitConfig struct {
	Enabled        bool `mapstructure:"rate_limit_enabled"`           // Enable/disable rate limiting
	RequestsPerMin int  `mapstructure:"rate_limit_requests_per_min"`  // Requests per minute allowed
	BurstSize      int  `mapstructure:"rate_limit_burst_size"`        // Max burst size
}

// ObservabilityConfig holds logging, metrics, and tracing settings.
//
// OBSERVABILITY EXPLANATION:
// In production, you need to understand what's happening in your system:
// - Logging: Text messages about what the application is doing
// - Metrics: Numerical measurements (requests/sec, error rate, response time)
// - Tracing: Following a request as it flows through multiple services
//
// These three pillars help you:
// - Debug issues in production
// - Understand performance characteristics
// - Detect problems before users notice them
type ObservabilityConfig struct {
	LogLevel       string `mapstructure:"log_level"`        // debug, info, warn, error, fatal
	MetricsEnabled bool   `mapstructure:"metrics_enabled"`  // Enable Prometheus metrics
	TracingEnabled bool   `mapstructure:"tracing_enabled"`  // Enable distributed tracing
	JaegerEndpoint string `mapstructure:"jaeger_endpoint"`  // Jaeger collector endpoint
}

// WorkerConfig holds background worker settings.
//
// WORKER PATTERN EXPLANATION:
// Workers are background processes that handle time-consuming tasks:
// - Image thumbnail generation
// - Video transcoding
// - File compression
// - Sending emails
//
// Instead of making the user wait, we queue the task and return immediately.
// Workers process tasks in the background.
//
// Concurrency controls how many tasks run simultaneously.
type WorkerConfig struct {
	Concurrency int `mapstructure:"worker_concurrency"` // Number of concurrent workers
}

// EmailConfig holds email notification settings.
type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`     // SMTP server hostname
	SMTPPort     int    `mapstructure:"smtp_port"`     // SMTP server port
	SMTPUsername string `mapstructure:"smtp_username"` // SMTP username
	SMTPPassword string `mapstructure:"smtp_password"` // SMTP password
	FromAddress  string `mapstructure:"email_from"`    // "From" email address
	SendGridKey  string `mapstructure:"sendgrid_api_key"` // SendGrid API key (alternative to SMTP)
}

// SecurityConfig holds security-related settings.
type SecurityConfig struct {
	CORSAllowedOrigins string `mapstructure:"cors_allowed_origins"` // Comma-separated list of allowed origins
	BcryptCost         int    `mapstructure:"bcrypt_cost"`          // Password hashing cost (4-31)
}

// ServicesConfig holds URLs for inter-service communication.
//
// MICROSERVICES COMMUNICATION:
// In a microservices architecture, services need to talk to each other.
// These URLs tell each service where to find the others.
// In Docker Compose, service names (like "auth-service") work as hostnames.
type ServicesConfig struct {
	AuthServiceURL         string `mapstructure:"auth_service_url"`
	FileServiceURL         string `mapstructure:"file_service_url"`
	MetadataServiceURL     string `mapstructure:"metadata_service_url"`
	ProcessingServiceURL   string `mapstructure:"processing_service_url"`
	VersioningServiceURL   string `mapstructure:"versioning_service_url"`
	NotificationServiceURL string `mapstructure:"notification_service_url"`
}

// =============================================================================
// CONFIGURATION LOADING FUNCTIONS
// =============================================================================

// Load reads configuration from environment variables and config files.
//
// FUNCTION EXPLANATION:
// In Go, functions are defined as:
//     func FunctionName(parameters) returnType { ... }
//
// This function takes a configPath string and returns two values:
// - *Config: A pointer to a Config struct (the asterisk * means pointer)
// - error: An error value (nil if no error occurred)
//
// ERROR HANDLING IN GO:
// Go uses explicit error handling instead of exceptions.
// Functions that can fail return an error as their last return value.
// The caller checks if error is nil (no error) or not nil (error occurred).
//
// POINTER EXPLANATION:
// A pointer holds the memory address of a value, not the value itself.
// We use pointers for large structs to avoid copying all the data.
// The & operator gets a pointer to a value.
// The * operator dereferences a pointer (gets the value it points to).
func Load(configPath string) (*Config, error) {
	// Create a new Viper instance
	// Viper is the configuration library that handles reading env vars and files
	v := viper.New()

	// Set default values
	// This ensures the application works even if some env vars are missing
	setDefaults(v)

	// If a config file path was provided, try to read it
	if configPath != "" {
		v.SetConfigFile(configPath)

		// ReadInConfig reads the config file
		// If it fails (file not found, invalid format, etc.), return the error
		if err := v.ReadInConfig(); err != nil {
			// In Go, we often return early on errors
			// This pattern is called "guard clause" - handle errors first, then happy path
			return nil, fmt.Errorf("failed to read config file: %w", err)
			// %w is a format verb that wraps the error, preserving the error chain
		}
	}

	// AutomaticEnv tells Viper to automatically read environment variables
	// Environment variables override config file values
	// This allows deploying the same code with different configurations
	v.AutomaticEnv()

	// Create a Config struct to hold the configuration
	// var declares a variable
	// cfg is the variable name
	// Config is the type
	var cfg Config

	// Unmarshal reads the configuration values into the cfg struct
	// It uses the mapstructure tags to know which config key goes to which field
	if err := v.Unmarshal(&cfg); err != nil {
		// The & operator creates a pointer to cfg
		// Unmarshal modifies the cfg struct, so it needs a pointer
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validation: ensure critical configuration values are set
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Return a pointer to the config and nil error (success)
	// The & operator creates a pointer to the cfg struct
	return &cfg, nil
}

// setDefaults sets default values for configuration.
//
// DEFAULT VALUES EXPLANATION:
// Defaults ensure the application has sensible values even if configuration
// is not provided. This makes local development easier and reduces errors.
//
// The second parameter to SetDefault is the default value.
// These can be overridden by environment variables or config files.
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("service_name", "file-storage-service")
	v.SetDefault("environment", "development")
	v.SetDefault("http_port", 8080)
	v.SetDefault("version", "1.0.0")
	v.SetDefault("shutdown_timeout", "30s")   // 30 seconds
	v.SetDefault("read_timeout", "15s")       // 15 seconds
	v.SetDefault("write_timeout", "15s")      // 15 seconds

	// Database defaults
	v.SetDefault("mongo_uri", "mongodb://localhost:27017")
	v.SetDefault("mongo_database", "file_storage")
	v.SetDefault("mongo_max_pool_size", 100)
	v.SetDefault("mongo_min_pool_size", 10)
	v.SetDefault("mongo_max_conn_idle_time", "30s")
	v.SetDefault("mongo_connect_timeout", "10s")

	// S3 defaults
	v.SetDefault("s3_endpoint", "")  // Empty means use AWS S3
	v.SetDefault("aws_region", "us-east-1")
	v.SetDefault("s3_bucket", "file-storage")
	v.SetDefault("s3_use_ssl", true)
	v.SetDefault("presign_url_expiry", "15m")  // 15 minutes
	v.SetDefault("chunk_size", 5242880)        // 5 MB in bytes
	v.SetDefault("max_file_size", 5368709120)  // 5 GB in bytes

	// RabbitMQ defaults
	v.SetDefault("rabbitmq_url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault("rabbitmq_exchange", "file-storage-exchange")
	v.SetDefault("rabbitmq_prefetch_count", 10)
	v.SetDefault("rabbitmq_reconnect_delay", "5s")

	// Redis defaults
	v.SetDefault("redis_host", "localhost")
	v.SetDefault("redis_port", 6379)
	v.SetDefault("redis_password", "")
	v.SetDefault("redis_db", 0)
	v.SetDefault("redis_max_retries", 3)
	v.SetDefault("redis_pool_size", 100)

	// JWT defaults
	// NOTE: In production, JWT_SECRET MUST be overridden with a secure random string!
	v.SetDefault("jwt_secret", "change-this-secret-in-production")
	v.SetDefault("jwt_expiry", "24h")     // 24 hours
	v.SetDefault("jwt_refresh_expiry", "168h")  // 7 days
	v.SetDefault("jwt_issuer", "file-storage-service")

	// Rate limiting defaults
	v.SetDefault("rate_limit_enabled", true)
	v.SetDefault("rate_limit_requests_per_min", 60)  // 1 request per second average
	v.SetDefault("rate_limit_burst_size", 10)

	// Observability defaults
	v.SetDefault("log_level", "info")
	v.SetDefault("metrics_enabled", true)
	v.SetDefault("tracing_enabled", true)
	v.SetDefault("jaeger_endpoint", "http://localhost:4318/v1/traces")

	// Worker defaults
	v.SetDefault("worker_concurrency", 10)

	// Email defaults
	v.SetDefault("smtp_host", "")
	v.SetDefault("smtp_port", 587)
	v.SetDefault("email_from", "noreply@file-storage.com")

	// Security defaults
	v.SetDefault("cors_allowed_origins", "*")  // In production, specify exact origins!
	v.SetDefault("bcrypt_cost", 12)  // Good balance of security and performance

	// Service URLs (for Docker Compose)
	v.SetDefault("auth_service_url", "http://auth-service:8081")
	v.SetDefault("file_service_url", "http://file-service:8082")
	v.SetDefault("metadata_service_url", "http://metadata-service:8083")
	v.SetDefault("processing_service_url", "http://processing-service:8084")
	v.SetDefault("versioning_service_url", "http://versioning-service:8085")
	v.SetDefault("notification_service_url", "http://notification-service:8086")
}

// Validate checks that required configuration values are properly set.
//
// METHOD EXPLANATION:
// This is a method on the Config struct (notice the receiver: (c *Config))
// Methods are functions that belong to a type.
// You call them like: cfg.Validate() instead of Validate(cfg)
//
// The receiver (c *Config) means this method can access the Config struct's fields.
// The * means it receives a pointer, so it can read (and modify, if needed) the struct.
func (c *Config) Validate() error {
	// Check MongoDB URI
	if c.Database.URI == "" {
		return fmt.Errorf("mongodb URI is required")
	}

	// Check database name
	if c.Database.Database == "" {
		return fmt.Errorf("mongodb database name is required")
	}

	// Check S3 bucket name
	if c.S3.Bucket == "" {
		return fmt.Errorf("S3 bucket name is required")
	}

	// Check S3 credentials
	if c.S3.AccessKeyID == "" || c.S3.SecretAccessKey == "" {
		return fmt.Errorf("S3 credentials (access key and secret key) are required")
	}

	// Check JWT secret
	if c.JWT.Secret == "" || c.JWT.Secret == "change-this-secret-in-production" {
		return fmt.Errorf("JWT secret must be set to a secure random string")
	}

	// Check RabbitMQ URL
	if c.RabbitMQ.URL == "" {
		return fmt.Errorf("RabbitMQ URL is required")
	}

	// Validation passed
	// In Go, returning nil for an error type means "no error"
	return nil
}

// GetRedisAddress returns the Redis address in host:port format.
//
// This is a helper method that formats the Redis connection string.
// Notice it returns a string, not an error - it can't fail.
func (c *Config) GetRedisAddress() string {
	// fmt.Sprintf is like printf - it formats a string
	// %s is a placeholder for string, %d is for integer
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// IsDevelopment returns true if running in development environment.
//
// Helper methods like this make code more readable:
// Instead of: if cfg.Server.Environment == "development"
// You can write: if cfg.IsDevelopment()
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================
// Here's how you would use this config package in your application:
//
//    package main
//
//    import (
//        "log"
//        "github.com/emaad/file-storage-service/pkg/config"
//    )
//
//    func main() {
//        // Load configuration
//        // Pass empty string to load from environment variables only
//        // Or pass a file path like "config.yaml" to load from file
//        cfg, err := config.Load("")
//        if err != nil {
//            log.Fatalf("Failed to load config: %v", err)
//        }
//
//        // Use configuration values
//        fmt.Printf("Starting server on port %d\n", cfg.Server.Port)
//        fmt.Printf("Connecting to MongoDB at %s\n", cfg.Database.URI)
//        fmt.Printf("Using S3 bucket: %s\n", cfg.S3.Bucket)
//
//        // Check environment
//        if cfg.IsDevelopment() {
//            fmt.Println("Running in development mode")
//        }
//    }
// =============================================================================
