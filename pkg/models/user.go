// Package models contains shared data structures used across services.
//
// LEARNING NOTES FOR GO BEGINNERS:
// =================================
// This file demonstrates:
// 1. Struct definitions for data modeling
// 2. BSON tags for MongoDB
// 3. JSON tags for API responses
// 4. The primitive.ObjectID type from MongoDB
// 5. Time handling in Go
// 6. Slice types for arrays
//
// DATA MODELING:
// These structs define the shape of our data - what fields each entity has
// and what types those fields are. Think of them as database schemas or
// TypeScript interfaces.
package models

import (
	"time" // For handling timestamps

	"go.mongodb.org/mongo-driver/bson/primitive" // MongoDB types
)

// =============================================================================
// USER MODEL
// =============================================================================

// User represents a user in the system.
//
// STRUCT TAGS EXPLANATION:
// Each field has tags (the strings in backticks after the type).
// Tags provide metadata about how to serialize/deserialize the field.
//
// BSON TAGS (for MongoDB):
//     `bson:"email"`      - Field name in MongoDB document
//     `bson:"_id,omitempty"` - "_id" is MongoDB's ID field, omitempty means skip if empty
//
// JSON TAGS (for API responses):
//     `json:"email"`      - Field name in JSON
//     `json:"-"`          - Don't include in JSON (used for sensitive data like passwords)
//     `json:"id,omitempty"` - "id" in JSON, omit if empty
//
// WHY DIFFERENT NAMES?
// MongoDB uses "_id", but in JSON APIs we typically use "id"
// MongoDB stores data as BSON (Binary JSON), APIs return JSON
//
// FIELD TYPES:
// primitive.ObjectID - MongoDB's unique identifier type (12 bytes, globally unique)
// string            - Text data
// int64             - 64-bit integer (for large numbers like bytes)
// time.Time         - Date and time
// []string          - Slice (dynamic array) of strings
// *time.Time        - Pointer to time.Time (allows nil for "not set")
type User struct {
	// ID is the unique identifier for the user
	// In MongoDB, every document has an "_id" field
	// primitive.ObjectID is a 12-byte identifier: timestamp(4) + random(5) + counter(3)
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// Email is the user's email address (unique)
	// We'll create a unique index on this field in MongoDB
	Email string `bson:"email" json:"email"`

	// PasswordHash stores the hashed password (NEVER store plain passwords!)
	// json:"-" means this field is NEVER included in JSON responses
	// This prevents accidentally leaking password hashes to clients
	PasswordHash string `bson:"password_hash" json:"-"`

	// Name is the user's display name
	Name string `bson:"name" json:"name"`

	// Role defines the user's permissions (e.g., "user", "admin", "premium")
	// We'll use this for Role-Based Access Control (RBAC)
	Role string `bson:"role" json:"role"`

	// StorageQuota is the maximum storage allowed for this user (in bytes)
	// int64 can hold values up to 9,223,372,036,854,775,807 (9+ exabytes!)
	// Example: 10GB = 10 * 1024 * 1024 * 1024 = 10,737,418,240 bytes
	StorageQuota int64 `bson:"storage_quota" json:"storage_quota"`

	// StorageUsed tracks how much storage the user is currently using (in bytes)
	// We'll update this whenever files are uploaded/deleted
	StorageUsed int64 `bson:"storage_used" json:"storage_used"`

	// APIKeys stores API keys for programmatic access
	// []APIKey is a slice (dynamic array) of APIKey structs
	// Users can have multiple API keys
	APIKeys []APIKey `bson:"api_keys,omitempty" json:"api_keys,omitempty"`

	// RateLimit stores rate limiting information
	// This is an embedded struct (composition pattern in Go)
	RateLimit RateLimitInfo `bson:"rate_limit" json:"rate_limit"`

	// CreatedAt stores when the user account was created
	// time.Time is Go's type for dates and times
	// In MongoDB, this is stored as an ISODate
	CreatedAt time.Time `bson:"created_at" json:"created_at"`

	// UpdatedAt stores when the user was last updated
	// We'll update this field whenever user data changes
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`

	// DeletedAt is used for soft deletes (marking as deleted without actually deleting)
	// *time.Time is a POINTER to time.Time
	// Pointers can be nil, meaning "not set"
	// If DeletedAt is nil: user is active
	// If DeletedAt has a value: user is soft-deleted
	// omitempty means skip this field if it's nil
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// APIKey represents an API key for programmatic access.
//
// EMBEDDED STRUCT:
// This struct is embedded in the User struct above.
// It's not stored in a separate MongoDB collection - it's part of the user document.
//
// USE CASE:
// API keys allow applications to authenticate without user credentials.
// Example: A mobile app uses an API key to access the user's files.
type APIKey struct {
	// Key is the actual API key string (hashed for security)
	// Like passwords, we hash API keys before storing them
	Key string `bson:"key" json:"-"` // Never expose in JSON!

	// Name is a human-readable label for this key
	// Example: "Mobile App", "Backup Script", "Integration Test"
	Name string `bson:"name" json:"name"`

	// CreatedAt stores when this key was created
	CreatedAt time.Time `bson:"created_at" json:"created_at"`

	// LastUsedAt stores when this key was last used
	// *time.Time because it might never have been used (nil)
	LastUsedAt *time.Time `bson:"last_used_at,omitempty" json:"last_used_at,omitempty"`

	// ExpiresAt is when this key expires (optional)
	// If nil, the key never expires
	ExpiresAt *time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// RateLimitInfo stores rate limiting data for the token bucket algorithm.
//
// TOKEN BUCKET ALGORITHM:
// Imagine a bucket that holds tokens:
// - Each request consumes one token
// - Tokens are added at a steady rate (refill)
// - If no tokens available, request is rejected
// - Allows bursts (bucket can hold multiple tokens)
//
// This struct stores the bucket state per user.
type RateLimitInfo struct {
	// RequestsPerMinute is how many requests are allowed per minute
	// This is the refill rate (tokens added per minute)
	RequestsPerMinute int `bson:"requests_per_minute" json:"requests_per_minute"`

	// Tokens is the current number of tokens in the bucket
	// This decreases with each request, increases over time
	// float64 allows fractional tokens (e.g., 5.5 tokens)
	Tokens float64 `bson:"tokens" json:"tokens"`

	// LastRefill is when we last refilled the bucket
	// We use this to calculate how many tokens to add
	LastRefill time.Time `bson:"last_refill" json:"last_refill"`
}

// =============================================================================
// USER METHODS
// =============================================================================
// Methods are functions that belong to a type.
// They're called like: user.HasPermission(...)
// =============================================================================

// IsActive returns true if the user is not soft-deleted.
//
// METHOD RECEIVER:
// (u *User) means this is a method on the User type
// u is the variable name for the User instance (like 'this' in other languages)
// * means pointer receiver (more efficient, doesn't copy the struct)
//
// RETURN TYPE:
// bool (true or false)
//
// LOGIC:
// If DeletedAt is nil (user is not deleted), return true
// If DeletedAt has a value (user is deleted), return false
func (u *User) IsActive() bool {
	return u.DeletedAt == nil
}

// HasStorageSpace checks if the user has enough storage space for a file.
//
// PARAMETERS:
// fileSize: Size of the file in bytes (int64)
//
// RETURN:
// bool: true if user has space, false otherwise
//
// LOGIC:
// Check if (current usage + new file size) <= quota
func (u *User) HasStorageSpace(fileSize int64) bool {
	return (u.StorageUsed + fileSize) <= u.StorageQuota
}

// IsAdmin returns true if the user has admin role.
//
// ROLE-BASED ACCESS CONTROL (RBAC):
// Different users have different roles with different permissions.
// Common roles: "user" (basic), "premium" (paid features), "admin" (full access)
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsPremium returns true if the user has a premium subscription.
func (u *User) IsPremium() bool {
	return u.Role == "premium" || u.Role == "admin"
}

// RemainingStorage returns how much storage space is left (in bytes).
//
// CALCULATION:
// Remaining = Quota - Used
// Returns 0 if quota is exceeded (no negative values)
func (u *User) RemainingStorage() int64 {
	remaining := u.StorageQuota - u.StorageUsed
	// If negative, return 0
	if remaining < 0 {
		return 0
	}
	return remaining
}

// AddStorage increases the storage used (call when uploading a file).
//
// USAGE:
//     user.AddStorage(fileSize)
//     // Save user to database
func (u *User) AddStorage(bytes int64) {
	u.StorageUsed += bytes
}

// RemoveStorage decreases the storage used (call when deleting a file).
//
// SAFETY:
// Ensures StorageUsed never goes negative
func (u *User) RemoveStorage(bytes int64) {
	u.StorageUsed -= bytes
	// Prevent negative storage
	if u.StorageUsed < 0 {
		u.StorageUsed = 0
	}
}

// =============================================================================
// VALIDATION
// =============================================================================

// Validate checks if the user data is valid.
//
// VALIDATION RULES:
// - Email is required
// - Name is required
// - Role is required and must be valid
// - StorageQuota must be positive
//
// RETURN:
// error if validation fails, nil if valid
//
// USAGE:
//     user := &User{Email: "test@example.com", ...}
//     if err := user.Validate(); err != nil {
//         return err  // Invalid user data
//     }
func (u *User) Validate() error {
	if u.Email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}

	if u.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}

	if u.Role == "" {
		return &ValidationError{Field: "role", Message: "role is required"}
	}

	// Validate role is one of the allowed values
	validRoles := map[string]bool{
		"user":    true,
		"premium": true,
		"admin":   true,
	}
	if !validRoles[u.Role] {
		return &ValidationError{
			Field:   "role",
			Message: "role must be one of: user, premium, admin",
		}
	}

	if u.StorageQuota <= 0 {
		return &ValidationError{
			Field:   "storage_quota",
			Message: "storage quota must be positive",
		}
	}

	return nil
}

// ValidationError represents a validation error.
//
// CUSTOM ERROR TYPE:
// This implements the error interface (has Error() method)
type ValidationError struct {
	Field   string // Which field failed validation
	Message string // Why it failed
}

// Error implements the error interface.
//
// ERROR INTERFACE:
// Any type with an Error() string method implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// NewUser creates a new user with default values.
//
// CONSTRUCTOR FUNCTION:
// In Go, we use functions named "NewXxx" to create and initialize structs.
// This is the Go equivalent of constructors in OOP languages.
//
// PARAMETERS:
// email, name, role: Required user information
//
// RETURN:
// *User: Pointer to a newly created User struct
//
// DEFAULT VALUES:
// - StorageQuota: 10GB for regular users
// - StorageUsed: 0 bytes
// - Role: "user" by default
// - CreatedAt/UpdatedAt: current time
// - RateLimit: 60 requests/minute with full bucket
func NewUser(email, name, role string) *User {
	now := time.Now()

	// Determine storage quota based on role
	var quota int64
	switch role {
	case "premium":
		quota = 100 * 1024 * 1024 * 1024 // 100 GB
	case "admin":
		quota = 1024 * 1024 * 1024 * 1024 // 1 TB
	default: // "user"
		quota = 10 * 1024 * 1024 * 1024 // 10 GB
	}

	// Determine rate limit based on role
	var requestsPerMin int
	switch role {
	case "premium":
		requestsPerMin = 300
	case "admin":
		requestsPerMin = 1000
	default: // "user"
		requestsPerMin = 60
	}

	// Return a pointer to a new User struct
	// The & operator creates a pointer
	return &User{
		ID:           primitive.NewObjectID(), // Generate new MongoDB ObjectID
		Email:        email,
		Name:         name,
		Role:         role,
		StorageQuota: quota,
		StorageUsed:  0,
		APIKeys:      []APIKey{},  // Empty slice of API keys
		RateLimit: RateLimitInfo{
			RequestsPerMinute: requestsPerMin,
			Tokens:            float64(requestsPerMin), // Start with full bucket
			LastRefill:        now,
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil, // nil means not deleted
	}
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================
//
// Creating a new user:
//
//     user := models.NewUser("john@example.com", "John Doe", "user")
//     user.PasswordHash = hashPassword("secret123")
//
// Checking storage:
//
//     if !user.HasStorageSpace(fileSize) {
//         return errors.New("QUOTA_EXCEEDED", "Storage quota exceeded", 403)
//     }
//     user.AddStorage(fileSize)
//
// Soft deleting:
//
//     now := time.Now()
//     user.DeletedAt = &now
//     // Save to database
//
// Checking if active:
//
//     if !user.IsActive() {
//         return errors.New("USER_DELETED", "User account is deleted", 403)
//     }
//
// MongoDB query example (not shown here, but for reference):
//
//     // Find active users
//     filter := bson.M{"deleted_at": nil}
//     cursor, err := collection.Find(ctx, filter)
//
//     // Find user by email
//     filter := bson.M{"email": "john@example.com"}
//     var user User
//     err := collection.FindOne(ctx, filter).Decode(&user)
//
// =============================================================================
