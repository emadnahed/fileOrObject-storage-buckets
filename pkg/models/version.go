// This file defines the FileVersion model for tracking file version history.
//
// LEARNING NOTES:
// ===============
// Demonstrates:
// 1. Versioning strategies (copy-on-write)
// 2. Historical data tracking
// 3. Content deduplication using checksums
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// =============================================================================
// FILE VERSION MODEL
// =============================================================================

// FileVersion represents a historical version of a file.
//
// VERSION CONTROL EXPLAINED:
// When a user updates a file, we create a new version instead of overwriting.
// This allows:
// - View version history
// - Restore previous versions
// - Compare versions
// - Audit who changed what when
//
// STORAGE STRATEGY (Copy-on-Write):
// 1. User uploads "document.pdf" (version 1)
// 2. User updates "document.pdf"
// 3. We move v1 to versions storage: "/versions/doc_v1.pdf"
// 4. We store the new content as current: "/files/document.pdf" (version 2)
// 5. Create a FileVersion record pointing to v1 in versions storage
//
// This ensures the current version is always fast to access.
type FileVersion struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// FileID is the current file this version belongs to
	// All versions of a file share the same FileID
	// This makes it easy to query all versions: find by file_id
	FileID primitive.ObjectID `bson:"file_id" json:"file_id"`

	// VersionNumber is the version number (1, 2, 3, ...)
	// Version 1 is the original upload
	// Each update increments the version number
	VersionNumber int `bson:"version_number" json:"version_number"`

	// S3Key is where this version is stored in S3
	// Format: "users/{user_id}/versions/{file_id}_v{version}.{ext}"
	// Example: "users/123/versions/abc_v1.pdf"
	//
	// WHY DIFFERENT FROM CURRENT FILE?
	// Current file: /files/document.pdf (always latest)
	// Old versions: /versions/document_v1.pdf, /versions/document_v2.pdf
	S3Key string `bson:"s3_key" json:"s3_key"`

	// S3Bucket and S3Region (same as file)
	S3Bucket string `bson:"s3_bucket" json:"s3_bucket"`
	S3Region string `bson:"s3_region" json:"s3_region"`

	// FileSize is the size of this version (in bytes)
	// Different versions can have different sizes
	FileSize int64 `bson:"file_size" json:"file_size"`

	// Checksum is the SHA-256 hash of this version's content
	//
	// DEDUPLICATION:
	// If two versions have the same checksum, they have identical content!
	// We can store one copy and have both versions point to it.
	// This saves storage space.
	//
	// Example:
	// v1: checksum = "abc123..." -> stored at /versions/file_v1.pdf
	// v2: checksum = "def456..." -> stored at /versions/file_v2.pdf
	// v3: checksum = "abc123..." -> same as v1! Point to /versions/file_v1.pdf
	//
	// Result: v3 doesn't consume additional storage
	Checksum string `bson:"checksum" json:"checksum"`

	// ChangesDescription is an optional description of what changed
	// Example: "Fixed typo in introduction", "Added new section on security"
	// Users can provide this when uploading a new version
	ChangesDescription string `bson:"changes_description,omitempty" json:"changes_description,omitempty"`

	// CreatedBy is who created this version
	// Usually the file owner, but could be a collaborator with write permission
	CreatedBy primitive.ObjectID `bson:"created_by" json:"created_by"`

	// CreatedAt is when this version was created
	CreatedAt time.Time `bson:"created_at" json:"created_at"`

	// RestoredAt is when this version was restored to become current
	// *time.Time because it might never have been restored (nil)
	//
	// RESTORE PROCESS:
	// 1. User selects version 3 to restore
	// 2. Current version (v5) is moved to versions storage
	// 3. Version 3 content is copied to current file location
	// 4. File.Version is set to 6 (new version)
	// 5. FileVersion record for v3 is updated: RestoredAt = now
	RestoredAt *time.Time `bson:"restored_at,omitempty" json:"restored_at,omitempty"`
}

// =============================================================================
// FILE VERSION METHODS
// =============================================================================

// WasRestored returns true if this version was restored.
//
// USAGE:
// Check if a version in the history was ever restored to current.
func (fv *FileVersion) WasRestored() bool {
	return fv.RestoredAt != nil
}

// IsDuplicate checks if this version has the same content as another.
//
// PARAMETERS:
// other: Another FileVersion to compare with
//
// RETURN:
// bool: true if checksums match (identical content)
//
// USE CASE:
// Before storing a new version, check if it's a duplicate.
// If yes, we can skip storing and just reference the existing version.
func (fv *FileVersion) IsDuplicate(other *FileVersion) bool {
	return fv.Checksum == other.Checksum
}

// GetStoragePath returns the S3 path for this version.
//
// HELPER METHOD:
// Centralizes the S3 path construction logic.
func (fv *FileVersion) GetStoragePath() string {
	return fv.S3Key
}

// =============================================================================
// CONSTRUCTOR FUNCTION
// =============================================================================

// NewFileVersion creates a new file version record.
//
// PARAMETERS:
// fileID:          The file this version belongs to
// versionNumber:   The version number (1, 2, 3, ...)
// s3Key:           Where this version is stored
// s3Bucket, s3Region: S3 location details
// fileSize:        Size in bytes
// checksum:        SHA-256 hash
// changesDesc:     Optional description
// createdBy:       Who created this version
//
// USAGE:
//     version := models.NewFileVersion(
//         fileID,
//         3,  // This is version 3
//         "users/123/versions/file_v3.pdf",
//         "file-storage",
//         "us-east-1",
//         1048576,  // 1MB
//         "e3b0c44...",
//         "Fixed critical bug",
//         userID,
//     )
func NewFileVersion(
	fileID primitive.ObjectID,
	versionNumber int,
	s3Key, s3Bucket, s3Region string,
	fileSize int64,
	checksum string,
	changesDesc string,
	createdBy primitive.ObjectID,
) *FileVersion {
	return &FileVersion{
		ID:                 primitive.NewObjectID(),
		FileID:             fileID,
		VersionNumber:      versionNumber,
		S3Key:              s3Key,
		S3Bucket:           s3Bucket,
		S3Region:           s3Region,
		FileSize:           fileSize,
		Checksum:           checksum,
		ChangesDescription: changesDesc,
		CreatedBy:          createdBy,
		CreatedAt:          time.Now(),
		RestoredAt:         nil,
	}
}

// =============================================================================
// USAGE EXAMPLES
// =============================================================================
//
// Creating a new version when file is updated:
//
//     // User uploads new version of existing file
//     currentFile, _ := fileRepo.GetByID(fileID)
//
//     // Create version record for the OLD content
//     oldVersionS3Key := fmt.Sprintf(
//         "users/%s/versions/%s_v%d.pdf",
//         userID.Hex(),
//         fileID.Hex(),
//         currentFile.Version,
//     )
//
//     version := models.NewFileVersion(
//         fileID,
//         currentFile.Version,  // Current version becomes historical
//         oldVersionS3Key,
//         currentFile.S3Bucket,
//         currentFile.S3Region,
//         currentFile.FileSize,
//         currentFile.Checksum,
//         "Automated version backup",
//         userID,
//     )
//     versionRepo.Create(version)
//
//     // Update current file with new content
//     currentFile.Version++  // Increment version
//     currentFile.Checksum = newChecksum
//     currentFile.FileSize = newFileSize
//     currentFile.UpdatedAt = time.Now()
//     fileRepo.Update(currentFile)
//
// Listing version history:
//
//     versions, err := versionRepo.FindByFileID(fileID)
//     for _, version := range versions {
//         fmt.Printf("Version %d: %s (created: %s)\n",
//             version.VersionNumber,
//             version.ChangesDescription,
//             version.CreatedAt.Format("2006-01-02"),
//         )
//     }
//
// Restoring a previous version:
//
//     // User wants to restore version 3
//     version, _ := versionRepo.GetByFileIDAndVersion(fileID, 3)
//
//     // Get current file
//     currentFile, _ := fileRepo.GetByID(fileID)
//
//     // Save current version to history first
//     newVersionRecord := models.NewFileVersion(
//         fileID,
//         currentFile.Version,
//         currentFile.S3Key,
//         currentFile.S3Bucket,
//         currentFile.S3Region,
//         currentFile.FileSize,
//         currentFile.Checksum,
//         "Pre-restore backup",
//         userID,
//     )
//     versionRepo.Create(newVersionRecord)
//
//     // Copy version 3 content to current file location
//     s3.CopyObject(version.S3Key, currentFile.S3Key)
//
//     // Update current file metadata
//     currentFile.Version++
//     currentFile.Checksum = version.Checksum
//     currentFile.FileSize = version.FileSize
//     currentFile.UpdatedAt = time.Now()
//     fileRepo.Update(currentFile)
//
//     // Mark version as restored
//     now := time.Now()
//     version.RestoredAt = &now
//     versionRepo.Update(version)
//
// Deduplication check:
//
//     // Before creating new version, check if content is identical
//     existingVersions, _ := versionRepo.FindByFileID(fileID)
//     newChecksum := calculateChecksum(newContent)
//
//     for _, existingVersion := range existingVersions {
//         if existingVersion.Checksum == newChecksum {
//             // Content is identical to an existing version!
//             // Option 1: Don't create new version, notify user
//             // Option 2: Create version record but point to existing S3 object
//             fmt.Println("No changes detected")
//             return
//         }
//     }
//
//     // Content is different, proceed with new version
//     version := models.NewFileVersion(...)
//
// =============================================================================
// VERSION RETENTION POLICIES
// =============================================================================
//
// To prevent unlimited version growth, implement retention policies:
//
// FREE USERS: Keep last 10 versions
//     versions, _ := versionRepo.FindByFileID(fileID)
//     if len(versions) > 10 {
//         // Delete oldest versions
//         oldestVersions := versions[10:]  // Keep first 10, delete rest
//         for _, old := range oldestVersions {
//             s3.DeleteObject(old.S3Key)
//             versionRepo.Delete(old.ID)
//         }
//     }
//
// PREMIUM USERS: Keep last 50 versions
// ADMIN USERS: Unlimited versions
//
// TIME-BASED RETENTION: Keep versions for 30 days
//     cutoff := time.Now().AddDate(0, 0, -30)  // 30 days ago
//     oldVersions := versionRepo.FindByFileIDOlderThan(fileID, cutoff)
//     for _, old := range oldVersions {
//         s3.DeleteObject(old.S3Key)
//         versionRepo.Delete(old.ID)
//     }
//
// =============================================================================
// MONGODB INDEXES
// =============================================================================
//
// Essential indexes for version operations:
//
// 1. Compound index on (file_id, version_number) for version lookups
//     db.file_versions.createIndex({ file_id: 1, version_number: -1 })
//
// 2. Index on checksum for deduplication queries
//     db.file_versions.createIndex({ checksum: 1 })
//
// 3. Index on created_at for retention policy cleanup
//     db.file_versions.createIndex({ created_at: 1 })
//
// =============================================================================
