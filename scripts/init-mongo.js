// MongoDB Initialization Script
//
// =============================================================================
// WHAT IS THIS FILE?
// =============================================================================
// This script runs automatically when MongoDB container starts for the first time.
// It creates the database, collections, and indexes we need.
//
// Docker automatically executes scripts in /docker-entrypoint-initdb.d/
// during container initialization.
//
// LEARNING NOTES:
// - MongoDB uses JavaScript for its shell (mongosh)
// - This script uses MongoDB shell commands
// - Collections are like tables in SQL databases
// - Indexes improve query performance
// =============================================================================

// Print a welcome message (visible in docker logs)
print('==================================================');
print('Initializing File Storage MongoDB Database');
print('==================================================');

// Switch to our application database
// If it doesn't exist, MongoDB creates it
// DATABASE NAMING:
// Using underscores (file_storage) is conventional for MongoDB database names
db = db.getSiblingDB('file_storage');

print('Creating database: file_storage');

// =============================================================================
// CREATE COLLECTIONS
// =============================================================================
// In MongoDB, collections are created automatically when you insert data.
// However, explicitly creating them allows us to set options and add indexes.
//
// WHY EXPLICIT CREATION?
// - Set validation rules
// - Configure collection options (capped collections, time series)
// - Create indexes at initialization time
// =============================================================================

// Create users collection
print('Creating users collection...');
db.createCollection('users', {
    // Validation rules ensure data quality
    validator: {
        $jsonSchema: {
            bsonType: 'object',
            required: ['email', 'password_hash', 'name', 'role'],
            properties: {
                email: {
                    bsonType: 'string',
                    description: 'must be a string and is required'
                },
                password_hash: {
                    bsonType: 'string',
                    description: 'must be a string and is required'
                },
                name: {
                    bsonType: 'string',
                    description: 'must be a string and is required'
                },
                role: {
                    enum: ['user', 'premium', 'admin'],
                    description: 'must be one of: user, premium, admin'
                },
                storage_quota: {
                    bsonType: 'long',
                    minimum: 0,
                    description: 'must be a positive number'
                },
                storage_used: {
                    bsonType: 'long',
                    minimum: 0,
                    description: 'must be a positive number'
                }
            }
        }
    }
});

// Create files collection
print('Creating files collection...');
db.createCollection('files');

// Create folders collection
print('Creating folders collection...');
db.createCollection('folders');

// Create file_versions collection
print('Creating file_versions collection...');
db.createCollection('file_versions');

// Create processing_jobs collection
print('Creating processing_jobs collection...');
db.createCollection('processing_jobs');

// Create notifications collection
print('Creating notifications collection...');
db.createCollection('notifications');

// Create activity_logs collection with TTL (Time To Live)
print('Creating activity_logs collection...');
db.createCollection('activity_logs');

// =============================================================================
// CREATE INDEXES
// =============================================================================
// Indexes dramatically improve query performance.
// Without indexes, MongoDB must scan every document (slow!).
// With indexes, MongoDB can quickly find matching documents.
//
// INDEX TYPES:
// - Single field: { field: 1 }  (1 = ascending, -1 = descending)
// - Compound: { field1: 1, field2: 1 }
// - Unique: { field: 1 }, { unique: true }
// - Text: { field: 'text' } for full-text search
// - TTL: { created_at: 1 }, { expireAfterSeconds: 3600 }
// =============================================================================

print('Creating indexes for users collection...');

// Unique index on email (prevents duplicate email addresses)
// UNIQUE INDEX: MongoDB ensures no two documents have the same email
db.users.createIndex(
    { email: 1 },
    { unique: true, name: 'email_unique_idx' }
);

// Index on created_at for sorting
db.users.createIndex(
    { created_at: -1 },
    { name: 'created_at_idx' }
);

// Compound index on role and created_at
// Useful for queries like "find all admin users, sorted by creation date"
db.users.createIndex(
    { role: 1, created_at: -1 },
    { name: 'role_created_idx' }
);

// Index on deleted_at for soft delete queries
// Sparse index: only indexes documents where deleted_at exists
db.users.createIndex(
    { deleted_at: 1 },
    { sparse: true, name: 'deleted_at_idx' }
);

// ---------------------------------------------------------------------------
// FILES COLLECTION INDEXES
// ---------------------------------------------------------------------------
print('Creating indexes for files collection...');

// Compound index on user_id and created_at
// MOST COMMON QUERY: "show my files, newest first"
db.files.createIndex(
    { user_id: 1, created_at: -1 },
    { name: 'user_files_idx' }
);

// Compound index on user_id and deleted_at
// QUERY: "show my active files" (deleted_at: null)
db.files.createIndex(
    { user_id: 1, deleted_at: 1 },
    { name: 'user_active_files_idx' }
);

// Unique index on s3_key (no two files with same S3 key)
db.files.createIndex(
    { s3_key: 1 },
    { unique: true, name: 's3_key_unique_idx' }
);

// Index on checksum for deduplication
// QUERY: "find files with same checksum as this one"
db.files.createIndex(
    { checksum: 1 },
    { name: 'checksum_idx' }
);

// Compound index for version queries
// QUERY: "find all versions of this file"
db.files.createIndex(
    { parent_file_id: 1, version: -1 },
    { name: 'file_versions_idx' }
);

// Index on folder_id for folder contents
// QUERY: "show all files in this folder"
db.files.createIndex(
    { folder_id: 1 },
    { name: 'folder_files_idx' }
);

// Index on shared_with.user_id for sharing queries
// QUERY: "find all files shared with me"
db.files.createIndex(
    { 'shared_with.user_id': 1 },
    { name: 'shared_files_idx' }
);

// Text index on file_name for search
// FULL-TEXT SEARCH: "find files with 'document' in the name"
db.files.createIndex(
    { file_name: 'text' },
    { name: 'file_name_text_idx' }
);

// Index on processing_status
// QUERY: "find all files waiting to be processed"
db.files.createIndex(
    { processing_status: 1 },
    { name: 'processing_status_idx' }
);

// ---------------------------------------------------------------------------
// FOLDERS COLLECTION INDEXES
// ---------------------------------------------------------------------------
print('Creating indexes for folders collection...');

// Compound index on user_id and parent_folder_id
// QUERY: "show folders in this location"
db.folders.createIndex(
    { user_id: 1, parent_folder_id: 1 },
    { name: 'user_folder_hierarchy_idx' }
);

// Index on path for prefix queries
// QUERY: "find all items under /Documents/Work"
db.folders.createIndex(
    { path: 1 },
    { name: 'folder_path_idx' }
);

// Compound index on user_id and path
db.folders.createIndex(
    { user_id: 1, path: 1 },
    { name: 'user_folder_path_idx' }
);

// Index on deleted_at for active folders
db.folders.createIndex(
    { deleted_at: 1 },
    { sparse: true, name: 'folder_deleted_idx' }
);

// ---------------------------------------------------------------------------
// FILE_VERSIONS COLLECTION INDEXES
// ---------------------------------------------------------------------------
print('Creating indexes for file_versions collection...');

// Compound index on file_id and version_number
// QUERY: "get version 3 of this file"
db.file_versions.createIndex(
    { file_id: 1, version_number: -1 },
    { name: 'file_version_idx' }
);

// Index on checksum for deduplication
db.file_versions.createIndex(
    { checksum: 1 },
    { name: 'version_checksum_idx' }
);

// Index on created_at for retention cleanup
// QUERY: "find versions older than 30 days"
db.file_versions.createIndex(
    { created_at: 1 },
    { name: 'version_created_idx' }
);

// ---------------------------------------------------------------------------
// PROCESSING_JOBS COLLECTION INDEXES
// ---------------------------------------------------------------------------
print('Creating indexes for processing_jobs collection...');

// Compound index on status and priority
// QUERY: "get next pending job with highest priority"
db.processing_jobs.createIndex(
    { status: 1, priority: -1 },
    { name: 'job_status_priority_idx' }
);

// Index on file_id
// QUERY: "get all jobs for this file"
db.processing_jobs.createIndex(
    { file_id: 1 },
    { name: 'job_file_idx' }
);

// Index on queued_at for FIFO processing
db.processing_jobs.createIndex(
    { queued_at: 1 },
    { name: 'job_queued_idx' }
);

// ---------------------------------------------------------------------------
// NOTIFICATIONS COLLECTION INDEXES
// ---------------------------------------------------------------------------
print('Creating indexes for notifications collection...');

// Compound index on user_id, read status, and created_at
// QUERY: "get my unread notifications, newest first"
db.notifications.createIndex(
    { user_id: 1, read: 1, created_at: -1 },
    { name: 'user_notifications_idx' }
);

// ---------------------------------------------------------------------------
// ACTIVITY_LOGS COLLECTION INDEXES (with TTL)
// ---------------------------------------------------------------------------
print('Creating indexes for activity_logs collection...');

// Index on user_id for user activity queries
db.activity_logs.createIndex(
    { user_id: 1, created_at: -1 },
    { name: 'user_activity_idx' }
);

// Index on file_id for file activity history
db.activity_logs.createIndex(
    { file_id: 1, created_at: -1 },
    { name: 'file_activity_idx' }
);

// TTL index: automatically delete logs older than 90 days
// AUTOMATIC CLEANUP: MongoDB deletes documents when created_at + 90 days < now
// 90 days = 90 * 24 * 60 * 60 = 7,776,000 seconds
db.activity_logs.createIndex(
    { created_at: 1 },
    { expireAfterSeconds: 7776000, name: 'activity_ttl_idx' }
);

// =============================================================================
// CREATE SAMPLE DATA (Optional - for development/testing)
// =============================================================================
// Uncomment this section if you want sample data for testing

/*
print('Inserting sample data...');

// Sample user
const sampleUser = {
    email: 'demo@example.com',
    password_hash: '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYzpLCL.s6G', // "password"
    name: 'Demo User',
    role: 'user',
    storage_quota: NumberLong(10737418240), // 10 GB
    storage_used: NumberLong(0),
    api_keys: [],
    rate_limit: {
        requests_per_minute: 60,
        tokens: 60,
        last_refill: new Date()
    },
    created_at: new Date(),
    updated_at: new Date()
};

db.users.insertOne(sampleUser);
print('Sample user created: demo@example.com / password');
*/

// =============================================================================
// VERIFICATION
// =============================================================================
print('\n==================================================');
print('Verifying database setup...');
print('==================================================');

// List all collections
print('\nCollections:');
db.getCollectionNames().forEach(function(collection) {
    print('  - ' + collection);
});

// Show index count for each collection
print('\nIndexes created:');
db.getCollectionNames().forEach(function(collection) {
    const indexes = db.getCollection(collection).getIndexes();
    print('  ' + collection + ': ' + indexes.length + ' indexes');
});

print('\n==================================================');
print('MongoDB initialization complete!');
print('==================================================');
print('');
print('Database: file_storage');
print('Connection string: mongodb://admin:changeme@localhost:27017');
print('');
print('You can now:');
print('1. Start your Go application');
print('2. Connect using MongoDB Compass');
print('3. Use mongosh: mongosh "mongodb://admin:changeme@localhost:27017/file_storage"');
print('');
print('==================================================');

// =============================================================================
// ADDITIONAL COMMANDS FOR DEVELOPMENT
// =============================================================================
//
// VIEW ALL USERS:
//     db.users.find().pretty()
//
// FIND FILES BY USER:
//     db.files.find({ user_id: ObjectId("...") }).pretty()
//
// COUNT DOCUMENTS:
//     db.files.countDocuments()
//
// LIST INDEXES:
//     db.files.getIndexes()
//
// EXPLAIN QUERY PERFORMANCE:
//     db.files.find({ user_id: ObjectId("...") }).explain("executionStats")
//
// DROP INDEX:
//     db.files.dropIndex("index_name")
//
// COMPACT COLLECTION (reclaim space):
//     db.runCommand({ compact: "files" })
//
// DATABASE STATS:
//     db.stats()
//
// COLLECTION STATS:
//     db.files.stats()
//
// =============================================================================
