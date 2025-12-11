// This file defines the Folder model for organizing files hierarchically.
//
// LEARNING NOTES:
// ===============
// Demonstrates:
// 1. Tree data structures in MongoDB (parent-child relationships)
// 2. Recursive operations (folder hierarchies)
// 3. Path management
package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// =============================================================================
// FOLDER MODEL
// =============================================================================

// Folder represents a directory that can contain files and other folders.
//
// HIERARCHICAL STRUCTURE:
// Folders form a tree structure:
//     Root
//     ├── Documents
//     │   ├── Work
//     │   └── Personal
//     └── Photos
//         └── Vacation
//
// Each folder (except root) has a ParentFolderID pointing to its parent.
// This creates a tree where we can navigate up and down the hierarchy.
type Folder struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// Name is the folder display name
	// Example: "Documents", "Photos", "Work"
	Name string `bson:"name" json:"name"`

	// UserID is who owns this folder
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`

	// ParentFolderID is the parent folder's ID
	// TREE STRUCTURE:
	// - nil for root-level folders (no parent)
	// - Points to parent folder ID for nested folders
	//
	// Example:
	// Root folder:      ParentFolderID = nil
	// /Documents:       ParentFolderID = nil (root level)
	// /Documents/Work:  ParentFolderID = Documents folder ID
	ParentFolderID *primitive.ObjectID `bson:"parent_folder_id,omitempty" json:"parent_folder_id,omitempty"`

	// Path is the full path from root to this folder
	// This makes it easy to display and query folder hierarchies
	//
	// Examples:
	// Root level folder:  "/Documents"
	// Nested folder:      "/Documents/Work"
	// Deeply nested:      "/Documents/Work/Projects/2024"
	//
	// WHY STORE PATH?
	// - Fast lookups (find all files in "/Documents/Work")
	// - Easy breadcrumb display in UI
	// - Simple prefix queries (find all items under "/Documents")
	Path string `bson:"path" json:"path"`

	// Color is an optional UI color for the folder
	// Example: "#FF5733", "blue", "green"
	// Some file systems let users color-code folders
	Color string `bson:"color,omitempty" json:"color,omitempty"`

	// Icon is an optional icon name for the folder
	// Example: "folder-documents", "folder-photos", "folder-work"
	Icon string `bson:"icon,omitempty" json:"icon,omitempty"`

	// SharedWith is a list of users this folder is shared with
	// Sharing a folder shares all files and subfolders within it
	SharedWith []SharedUser `bson:"shared_with,omitempty" json:"shared_with,omitempty"`

	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`

	// DeletedAt is for soft delete
	// When a folder is deleted, all files and subfolders are also marked deleted
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// =============================================================================
// FOLDER METHODS
// =============================================================================

// IsActive returns true if the folder is not soft-deleted.
func (f *Folder) IsActive() bool {
	return f.DeletedAt == nil
}

// IsRootLevel returns true if this is a root-level folder.
//
// ROOT LEVEL:
// Folders with no parent (ParentFolderID is nil) are at root level.
// Example: /Documents, /Photos
func (f *Folder) IsRootLevel() bool {
	return f.ParentFolderID == nil
}

// GetDepth returns the nesting depth of the folder.
//
// DEPTH CALCULATION:
// Count the number of "/" in the path (excluding leading "/")
//
// Examples:
// "/Documents"           -> depth 1
// "/Documents/Work"      -> depth 2
// "/Documents/Work/2024" -> depth 3
//
// RETURN:
// int: The folder depth
func (f *Folder) GetDepth() int {
	// Remove leading and trailing slashes
	// strings.Trim removes characters from both ends
	trimmed := strings.Trim(f.Path, "/")

	// If empty string (root), depth is 0
	if trimmed == "" {
		return 0
	}

	// Count slashes and add 1
	// strings.Count counts occurrences of a substring
	// Example: "Documents/Work" has 1 slash, depth is 2
	return strings.Count(trimmed, "/") + 1
}

// GetParentPath returns the path of the parent folder.
//
// PATH MANIPULATION:
// Given "/Documents/Work/Projects", return "/Documents/Work"
//
// RETURN:
// string: Parent folder path
// If already at root, returns ""
func (f *Folder) GetParentPath() string {
	// Find the last occurrence of "/"
	// strings.LastIndex returns the index, or -1 if not found
	lastSlash := strings.LastIndex(f.Path, "/")

	if lastSlash <= 0 {
		return "" // Already at root or invalid path
	}

	// Return everything before the last slash
	// Example: "/Documents/Work/Projects"[:15] = "/Documents/Work"
	return f.Path[:lastSlash]
}

// IsSharedWith checks if the folder is shared with a specific user.
func (f *Folder) IsSharedWith(userID primitive.ObjectID) bool {
	for _, shared := range f.SharedWith {
		if shared.UserID.Hex() == userID.Hex() {
			return true
		}
	}
	return false
}

// GetPermission returns the permission level for a specific user.
func (f *Folder) GetPermission(userID primitive.ObjectID) (FilePermission, bool) {
	// Check if user is the owner
	if f.UserID.Hex() == userID.Hex() {
		return PermissionAdmin, true
	}

	// Check shared permissions
	for _, shared := range f.SharedWith {
		if shared.UserID.Hex() == userID.Hex() {
			return shared.Permission, true
		}
	}

	return "", false
}

// =============================================================================
// CONSTRUCTOR FUNCTION
// =============================================================================

// NewFolder creates a new folder.
//
// PARAMETERS:
// userID:         Who owns the folder
// name:           Folder display name
// parentFolderID: Parent folder ID (nil for root level)
// parentPath:     Parent folder path (empty for root level)
//
// PATH CONSTRUCTION:
// If parent path is empty: path = "/name"
// If parent path exists: path = "/parent/name"
func NewFolder(userID primitive.ObjectID, name string, parentFolderID *primitive.ObjectID, parentPath string) *Folder {
	now := time.Now()

	// Construct the full path
	var path string
	if parentPath == "" {
		// Root level folder
		path = "/" + name
	} else {
		// Nested folder
		path = parentPath + "/" + name
	}

	return &Folder{
		ID:             primitive.NewObjectID(),
		Name:           name,
		UserID:         userID,
		ParentFolderID: parentFolderID,
		Path:           path,
		SharedWith:     []SharedUser{},
		CreatedAt:      now,
		UpdatedAt:      now,
		DeletedAt:      nil,
	}
}

// NewRootFolder creates a root-level folder.
//
// CONVENIENCE FUNCTION:
// Makes it easier to create root folders without passing nil and empty string.
func NewRootFolder(userID primitive.ObjectID, name string) *Folder {
	return NewFolder(userID, name, nil, "")
}

// =============================================================================
// USAGE EXAMPLES
// =============================================================================
//
// Creating a root folder:
//
//     folder := models.NewRootFolder(userID, "Documents")
//     // Result: Path = "/Documents", ParentFolderID = nil
//
// Creating a nested folder:
//
//     parentFolder := models.NewRootFolder(userID, "Documents")
//     // Save parentFolder to get its ID
//
//     workFolder := models.NewFolder(
//         userID,
//         "Work",
//         &parentFolder.ID,
//         parentFolder.Path,
//     )
//     // Result: Path = "/Documents/Work"
//
// Finding all items in a folder (MongoDB query):
//
//     // Find all files in "/Documents" folder
//     filter := bson.M{
//         "user_id": userID,
//         "file_path": bson.M{"$regex": "^/Documents/"},
//         "deleted_at": nil,
//     }
//     cursor, err := filesCollection.Find(ctx, filter)
//
//     // Find all subfolders in "/Documents"
//     filter := bson.M{
//         "user_id": userID,
//         "path": bson.M{"$regex": "^/Documents/"},
//         "deleted_at": nil,
//     }
//     cursor, err := foldersCollection.Find(ctx, filter)
//
// Moving a folder (requires updating all child paths):
//
//     // When moving "/Documents/Work" to "/Archive/Work"
//     // Must update:
//     // 1. The folder's path
//     // 2. All descendant folders' paths
//     // 3. All files' paths within this folder tree
//
//     oldPrefix := "/Documents/Work"
//     newPrefix := "/Archive/Work"
//
//     // Update all descendant folder paths
//     filter := bson.M{"path": bson.M{"$regex": "^" + oldPrefix}}
//     // Use $replaceRoot or update operations to replace path prefix
//
// Deleting a folder (cascading delete):
//
//     now := time.Now()
//     folder.DeletedAt = &now
//
//     // Also soft-delete all descendant folders
//     filter := bson.M{"path": bson.M{"$regex": "^" + folder.Path + "/"}}
//     update := bson.M{"$set": bson.M{"deleted_at": now}}
//     foldersCollection.UpdateMany(ctx, filter, update)
//
//     // And all files in this folder tree
//     filter = bson.M{"file_path": bson.M{"$regex": "^" + folder.Path + "/"}}
//     update = bson.M{"$set": bson.M{"deleted_at": now}}
//     filesCollection.UpdateMany(ctx, filter, update)
//
// =============================================================================
// PERFORMANCE CONSIDERATIONS
// =============================================================================
//
// INDEXES:
// Essential indexes for folder operations:
//
// 1. Compound index on (user_id, path) for folder lookups
// 2. Index on (user_id, parent_folder_id) for listing folder contents
// 3. Index on path with regex support for prefix queries
//
// MongoDB index creation:
//     db.folders.createIndex({ user_id: 1, path: 1 })
//     db.folders.createIndex({ user_id: 1, parent_folder_id: 1 })
//     db.folders.createIndex({ path: 1 })
//
// PATH vs PARENT_FOLDER_ID:
// We store both for different use cases:
// - path: Fast prefix queries ("show everything under /Documents")
// - parent_folder_id: Fast parent lookup ("show direct children of this folder")
//
// =============================================================================
