// This file defines the File model - the core entity of our file storage system.
//
// LEARNING NOTES:
// ===============
// This demonstrates:
// 1. Complex struct modeling with nested structures
// 2. Enum-like constants in Go
// 3. Embedded structs (composition)
// 4. File metadata management
// 5. S3 multipart upload tracking
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// =============================================================================
// FILE STATUS CONSTANTS
// =============================================================================
// In Go, we don't have enums like in other languages.
// Instead, we define constants of a custom type.
//
// const declares constants (values that never change)
// The = iota pattern creates auto-incrementing values (0, 1, 2, ...)
// This is similar to enums in other languages.
// =============================================================================

// ProcessingStatus represents the processing state of a file.
//
// TYPE DEFINITION:
// type ProcessingStatus string creates a new type based on string
// This provides type safety - you can't accidentally use a regular string
// where a ProcessingStatus is expected.
//
// WHY USE A CUSTOM TYPE?
// - Type safety (compiler catches mistakes)
// - Self-documenting code (ProcessingStatus vs string)
// - Easy to add methods later
type ProcessingStatus string

// Processing status constants
//
// CONSTANT DECLARATION:
// const ( ... ) declares multiple constants at once
// We're defining all possible values for ProcessingStatus
const (
	ProcessingPending   ProcessingStatus = "pending"    // Waiting to be processed
	ProcessingInProgress ProcessingStatus = "processing" // Currently being processed
	ProcessingCompleted ProcessingStatus = "completed"   // Processing finished
	ProcessingFailed    ProcessingStatus = "failed"      // Processing failed
)

// UploadStatus represents the upload state for chunked uploads.
type UploadStatus string

// Upload status constants
const (
	UploadInitiated  UploadStatus = "initiated"   // Multipart upload initiated
	UploadInProgress UploadStatus = "in_progress" // Chunks being uploaded
	UploadCompleted  UploadStatus = "completed"   // All chunks uploaded
	UploadAborted    UploadStatus = "aborted"     // Upload cancelled
)

// FilePermission represents sharing permissions.
type FilePermission string

// Permission constants
const (
	PermissionRead  FilePermission = "read"  // Can view/download
	PermissionWrite FilePermission = "write" // Can modify
	PermissionAdmin FilePermission = "admin" // Can delete/share
)

// =============================================================================
// FILE MODEL
// =============================================================================

// File represents a file in the storage system.
//
// THIS IS A COMPLEX STRUCT:
// It demonstrates many Go concepts:
// - Multiple data types (primitives, slices, pointers, nested structs)
// - Versioning (parent_file_id, version fields)
// - Upload state tracking (for chunked uploads)
// - Sharing and permissions
// - Processing status (thumbnails, compression)
// - Timestamps (created, updated, deleted, accessed)
type File struct {
	// -------------------------------------------------------------------------
	// BASIC IDENTIFICATION
	// -------------------------------------------------------------------------

	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// UserID is who owns this file
	// All files belong to a user
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`

	// FileName is the display name (what the user sees)
	// Example: "My Document.pdf", "photo.jpg"
	FileName string `bson:"file_name" json:"file_name"`

	// FolderID is the parent folder (optional)
	// nil means file is in root folder
	FolderID *primitive.ObjectID `bson:"folder_id,omitempty" json:"folder_id,omitempty"`

	// FilePath is the virtual path in the user's directory structure
	// Example: "/Documents/Work/Report.pdf"
	// This makes it easy to display file hierarchy
	FilePath string `bson:"file_path" json:"file_path"`

	// -------------------------------------------------------------------------
	// S3 STORAGE INFORMATION
	// -------------------------------------------------------------------------
	// These fields track where the file is stored in S3

	// S3Key is the actual key (path) in the S3 bucket
	// Format: "users/{user_id}/files/{file_id}.{ext}"
	// Example: "users/507f1f77bcf86cd799439011/files/image123.jpg"
	// This is unique - no two files have the same S3 key
	S3Key string `bson:"s3_key" json:"s3_key"`

	// S3Bucket is the name of the S3 bucket
	// Example: "file-storage-prod"
	S3Bucket string `bson:"s3_bucket" json:"s3_bucket"`

	// S3Region is the AWS region where the bucket is located
	// Example: "us-east-1", "eu-west-1"
	S3Region string `bson:"s3_region" json:"s3_region"`

	// -------------------------------------------------------------------------
	// FILE METADATA
	// -------------------------------------------------------------------------

	// FileSize is the size in bytes
	// int64 can handle files up to 9+ exabytes (very large!)
	FileSize int64 `bson:"file_size" json:"file_size"`

	// MimeType is the file's content type
	// Examples: "image/jpeg", "application/pdf", "video/mp4"
	// We use this to determine how to process the file
	MimeType string `bson:"mime_type" json:"mime_type"`

	// Checksum is the SHA-256 hash of the file content
	// This is used for:
	// - Deduplication (two files with same hash = same content)
	// - Integrity verification (detect if file was corrupted)
	// - Fast comparison (hash is faster than comparing bytes)
	// Example: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	Checksum string `bson:"checksum" json:"checksum"`

	// -------------------------------------------------------------------------
	// VERSIONING
	// -------------------------------------------------------------------------
	// Files can have multiple versions (like Google Docs version history)

	// Version is the current version number
	// Starts at 1, increments with each update
	Version int `bson:"version" json:"version"`

	// IsLatest indicates if this is the current version
	// Only one version of a file should have IsLatest=true
	// All older versions have IsLatest=false
	IsLatest bool `bson:"is_latest" json:"is_latest"`

	// ParentFileID links to the previous version
	// nil for the first version
	// If file has versions, older versions point to their parent
	// This creates a version chain: v1 <- v2 <- v3 (current)
	ParentFileID *primitive.ObjectID `bson:"parent_file_id,omitempty" json:"parent_file_id,omitempty"`

	// -------------------------------------------------------------------------
	// PROCESSING STATUS
	// -------------------------------------------------------------------------
	// Background workers process files (thumbnails, compression, etc.)

	// ProcessingStatus tracks if background processing is done
	// See ProcessingStatus constants above
	ProcessingStatus ProcessingStatus `bson:"processing_status" json:"processing_status"`

	// ThumbnailURL is the URL of the generated thumbnail (for images/videos)
	// nil if no thumbnail exists
	ThumbnailURL *string `bson:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`

	// CompressedURL is the URL of the compressed version
	// nil if not compressed or compression not applicable
	CompressedURL *string `bson:"compressed_url,omitempty" json:"compressed_url,omitempty"`

	// Metadata stores file-specific metadata
	// For images: width, height, format
	// For videos: duration, codec, resolution
	// For documents: page count, author
	// This is a flexible map that can store any key-value pairs
	// interface{} means "any type"
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`

	// -------------------------------------------------------------------------
	// CHUNKED UPLOAD TRACKING
	// -------------------------------------------------------------------------
	// For large files, we use multipart upload (split into chunks)

	// UploadID is the S3 multipart upload ID
	// S3 returns this when we initiate a multipart upload
	// We use it to complete/abort the upload
	UploadID string `bson:"upload_id,omitempty" json:"upload_id,omitempty"`

	// Chunks tracks which chunks have been uploaded
	// []UploadChunk is a slice of UploadChunk structs
	// When all chunks are uploaded, we complete the multipart upload
	Chunks []UploadChunk `bson:"chunks,omitempty" json:"chunks,omitempty"`

	// UploadStatus tracks the upload progress
	// See UploadStatus constants above
	UploadStatus UploadStatus `bson:"upload_status" json:"upload_status"`

	// -------------------------------------------------------------------------
	// SHARING AND PERMISSIONS
	// -------------------------------------------------------------------------

	// OwnerID is who owns this file
	// Usually same as UserID, but can differ if file is shared and copied
	OwnerID primitive.ObjectID `bson:"owner_id" json:"owner_id"`

	// SharedWith is a list of users this file is shared with
	// []SharedUser is a slice of SharedUser structs
	// Empty slice means file is not shared
	SharedWith []SharedUser `bson:"shared_with,omitempty" json:"shared_with,omitempty"`

	// IsPublic indicates if file has a public link
	// If true, anyone with the link can access the file
	IsPublic bool `bson:"is_public" json:"is_public"`

	// PublicURL is the public access URL (if IsPublic=true)
	// nil if file is not public
	PublicURL *string `bson:"public_url,omitempty" json:"public_url,omitempty"`

	// -------------------------------------------------------------------------
	// TIMESTAMPS
	// -------------------------------------------------------------------------

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`

	// DeletedAt is for soft delete (nil = not deleted)
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`

	// LastAccessedAt tracks when file was last viewed/downloaded
	// Useful for analytics and auto-archiving old files
	LastAccessedAt *time.Time `bson:"last_accessed_at,omitempty" json:"last_accessed_at,omitempty"`
}

// =============================================================================
// EMBEDDED STRUCTS
// =============================================================================

// UploadChunk represents a single chunk in a multipart upload.
//
// MULTIPART UPLOAD EXPLAINED:
// Large files (>100MB) are split into chunks and uploaded separately.
// Benefits:
// - Resume failed uploads (re-upload only failed chunks)
// - Parallel uploads (upload multiple chunks simultaneously)
// - Better reliability (network issues affect only one chunk)
//
// S3 multipart upload process:
// 1. Initiate multipart upload -> get UploadID
// 2. Upload each chunk -> get ETag for each
// 3. Complete multipart upload with list of (PartNumber, ETag) pairs
type UploadChunk struct {
	// PartNumber is the chunk number (1-based: 1, 2, 3, ...)
	// S3 requires this to reassemble chunks in correct order
	PartNumber int `bson:"part_number" json:"part_number"`

	// ETag is the entity tag returned by S3
	// It's a hash of the chunk data
	// We need this to complete the multipart upload
	// Example: "\"33a64df551425fcc55e4d42a148795d9f25f89d4\""
	ETag string `bson:"etag" json:"etag"`

	// UploadedAt is when this chunk was successfully uploaded
	UploadedAt time.Time `bson:"uploaded_at" json:"uploaded_at"`

	// Size is the chunk size in bytes
	// Useful for progress tracking
	Size int64 `bson:"size" json:"size"`
}

// SharedUser represents a user with whom a file is shared.
//
// SHARING MODEL:
// Files can be shared with specific users with different permission levels.
// Example: Share document with coworker (read permission)
//          Share project folder with team lead (write permission)
type SharedUser struct {
	// UserID is who the file is shared with
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`

	// Permission defines what they can do
	// See FilePermission constants above
	Permission FilePermission `bson:"permission" json:"permission"`

	// SharedAt is when the file was shared
	SharedAt time.Time `bson:"shared_at" json:"shared_at"`

	// SharedBy is who shared the file
	SharedBy primitive.ObjectID `bson:"shared_by" json:"shared_by"`
}

// =============================================================================
// FILE METHODS
// =============================================================================

// IsActive returns true if the file is not soft-deleted.
func (f *File) IsActive() bool {
	return f.DeletedAt == nil
}

// IsImage returns true if the file is an image.
//
// MIME TYPE CHECKING:
// Common image MIME types:
// - image/jpeg
// - image/png
// - image/gif
// - image/webp
//
// We check if the MIME type starts with "image/"
func (f *File) IsImage() bool {
	return len(f.MimeType) >= 6 && f.MimeType[:6] == "image/"
}

// IsVideo returns true if the file is a video.
func (f *File) IsVideo() bool {
	return len(f.MimeType) >= 6 && f.MimeType[:6] == "video/"
}

// IsDocument returns true if the file is a document.
func (f *File) IsDocument() bool {
	// Common document MIME types
	docTypes := map[string]bool{
		"application/pdf":                                                true,
		"application/msword":                                             true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel":                                       true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/plain":                                                     true,
	}
	return docTypes[f.MimeType]
}

// CanCompress returns true if the file type supports compression.
//
// COMPRESSION LOGIC:
// - Don't compress already compressed formats (jpg, mp4, zip)
// - Compress text-based formats (txt, json, csv)
// - Compress documents (pdf, docx)
func (f *File) CanCompress() bool {
	// Already compressed formats - don't compress again
	compressedFormats := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true, // PNG is already compressed
		"video/mp4":  true,
		"video/x-matroska": true,
		"application/zip":  true,
		"application/x-rar-compressed": true,
	}

	if compressedFormats[f.MimeType] {
		return false
	}

	// Text-based and documents benefit from compression
	return f.IsDocument() || f.MimeType == "text/plain" || f.MimeType == "application/json"
}

// IsSharedWith checks if the file is shared with a specific user.
//
// PARAMETERS:
// userID: The user to check
//
// RETURN:
// bool: true if shared with this user
//
// ALGORITHM:
// Loop through SharedWith slice, check if any entry matches userID
func (f *File) IsSharedWith(userID primitive.ObjectID) bool {
	// range loops over slices
	// shared is each element in the slice
	for _, shared := range f.SharedWith {
		// Compare ObjectIDs using Hex() method (converts to string)
		if shared.UserID.Hex() == userID.Hex() {
			return true
		}
	}
	return false
}

// GetPermission returns the permission level for a specific user.
//
// RETURN:
// FilePermission: The permission level
// bool: true if user has any permission, false otherwise
func (f *File) GetPermission(userID primitive.ObjectID) (FilePermission, bool) {
	// Check if user is the owner
	if f.OwnerID.Hex() == userID.Hex() {
		return PermissionAdmin, true // Owners have admin permission
	}

	// Check shared permissions
	for _, shared := range f.SharedWith {
		if shared.UserID.Hex() == userID.Hex() {
			return shared.Permission, true
		}
	}

	// Not shared with this user
	return "", false
}

// AddChunk adds an uploaded chunk to the tracking list.
//
// USAGE:
//     file.AddChunk(1, "etag123", 5242880)  // Chunk 1, 5MB
func (f *File) AddChunk(partNumber int, etag string, size int64) {
	chunk := UploadChunk{
		PartNumber: partNumber,
		ETag:       etag,
		UploadedAt: time.Now(),
		Size:       size,
	}
	f.Chunks = append(f.Chunks, chunk)  // append adds to slice
}

// AllChunksUploaded checks if all expected chunks are uploaded.
//
// PARAMETERS:
// expectedChunks: How many chunks we expect
//
// RETURN:
// bool: true if all chunks uploaded
func (f *File) AllChunksUploaded(expectedChunks int) bool {
	return len(f.Chunks) == expectedChunks
}

// =============================================================================
// CONSTRUCTOR FUNCTION
// =============================================================================

// NewFile creates a new file with default values.
//
// PARAMETERS:
// userID:   Who owns the file
// fileName: Display name
// fileSize: Size in bytes
// mimeType: Content type
// s3Key:    S3 object key
// s3Bucket: S3 bucket name
// s3Region: AWS region
// checksum: SHA-256 hash
func NewFile(userID primitive.ObjectID, fileName string, fileSize int64, mimeType, s3Key, s3Bucket, s3Region, checksum string) *File {
	now := time.Now()

	return &File{
		ID:               primitive.NewObjectID(),
		UserID:           userID,
		FileName:         fileName,
		FolderID:         nil, // Root folder
		FilePath:         "/" + fileName,
		S3Key:            s3Key,
		S3Bucket:         s3Bucket,
		S3Region:         s3Region,
		FileSize:         fileSize,
		MimeType:         mimeType,
		Checksum:         checksum,
		Version:          1,
		IsLatest:         true,
		ParentFileID:     nil,
		ProcessingStatus: ProcessingPending,
		ThumbnailURL:     nil,
		CompressedURL:    nil,
		Metadata:         make(map[string]interface{}),  // Empty map
		Chunks:           []UploadChunk{},                // Empty slice
		UploadStatus:     UploadCompleted,
		OwnerID:          userID,
		SharedWith:       []SharedUser{},  // Not shared
		IsPublic:         false,
		PublicURL:        nil,
		CreatedAt:        now,
		UpdatedAt:        now,
		DeletedAt:        nil,
		LastAccessedAt:   nil,
	}
}

// NewFileForChunkedUpload creates a file for chunked upload.
func NewFileForChunkedUpload(userID primitive.ObjectID, fileName string, fileSize int64, mimeType, s3Key, s3Bucket, s3Region, uploadID string) *File {
	file := NewFile(userID, fileName, fileSize, mimeType, s3Key, s3Bucket, s3Region, "")
	file.UploadID = uploadID
	file.UploadStatus = UploadInitiated
	return file
}

// =============================================================================
// USAGE EXAMPLES
// =============================================================================
//
// Creating a file:
//
//     file := models.NewFile(
//         userID,
//         "document.pdf",
//         1048576,  // 1MB
//         "application/pdf",
//         "users/123/files/doc.pdf",
//         "file-storage",
//         "us-east-1",
//         "e3b0c44...",
//     )
//
// Chunked upload:
//
//     file := models.NewFileForChunkedUpload(userID, "large.zip", ...)
//     file.AddChunk(1, "etag1", 5242880)
//     file.AddChunk(2, "etag2", 5242880)
//     if file.AllChunksUploaded(10) {
//         file.UploadStatus = models.UploadCompleted
//     }
//
// Sharing:
//
//     sharedUser := models.SharedUser{
//         UserID:     friendID,
//         Permission: models.PermissionRead,
//         SharedAt:   time.Now(),
//         SharedBy:   ownerID,
//     }
//     file.SharedWith = append(file.SharedWith, sharedUser)
//
// Checking permissions:
//
//     permission, hasAccess := file.GetPermission(userID)
//     if !hasAccess {
//         return errors.ErrForbidden
//     }
//     if permission == models.PermissionRead {
//         // User can only read, not modify
//     }
//
// =============================================================================
