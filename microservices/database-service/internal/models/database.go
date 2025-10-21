// Package models provides data models for the Database Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Database represents a database instance.
type Database struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"not null;uniqueIndex" json:"name"`
	Type        string         `gorm:"not null" json:"type"` // postgresql, mysql, sqlite, mongodb
	Host        string         `gorm:"not null" json:"host"`
	Port        int            `gorm:"not null" json:"port"`
	User        string         `gorm:"not null" json:"user"`
	Password    string         `gorm:"not null" json:"password"`
	Database    string         `gorm:"not null" json:"database"`
	SSLMode     string         `json:"ssl_mode"`
	MaxConns    int            `json:"max_conns"`
	MinConns    int            `json:"min_conns"`
	MaxIdle     int            `json:"max_idle"`
	MaxLifetime int            `json:"max_lifetime"`
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, maintenance
	Metadata    string         `gorm:"type:text" json:"metadata"`    // JSON string for additional data
}

// Table represents a database table.
type Table struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DatabaseID uint           `gorm:"not null;index" json:"database_id"`
	Database   Database       `gorm:"foreignKey:DatabaseID" json:"database"`
	Name       string         `gorm:"not null;index" json:"name"`
	Schema     string         `json:"schema"`
	Engine     string         `json:"engine"`
	Charset    string         `json:"charset"`
	Collation  string         `json:"collation"`
	RowCount   int64          `json:"row_count"`
	SizeBytes  int64          `json:"size_bytes"`
	Status     string         `gorm:"default:active" json:"status"` // active, inactive, maintenance
	Metadata   string         `gorm:"type:text" json:"metadata"`    // JSON string for additional data
}

// Column represents a table column.
type Column struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TableID     uint           `gorm:"not null;index" json:"table_id"`
	Table       Table          `gorm:"foreignKey:TableID" json:"table"`
	Name        string         `gorm:"not null;index" json:"name"`
	Type        string         `gorm:"not null" json:"type"`
	Length      int            `json:"length"`
	Precision   int            `json:"precision"`
	Scale       int            `json:"scale"`
	Nullable    bool           `json:"nullable"`
	Default     string         `json:"default"`
	PrimaryKey  bool           `json:"primary_key"`
	Unique      bool           `json:"unique"`
	Indexed     bool           `json:"indexed"`
	Description string         `json:"description"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Index represents a database index.
type Index struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TableID   uint           `gorm:"not null;index" json:"table_id"`
	Table     Table          `gorm:"foreignKey:TableID" json:"table"`
	Name      string         `gorm:"not null;index" json:"name"`
	Type      string         `gorm:"not null" json:"type"`    // primary, unique, index, fulltext
	Columns   string         `gorm:"not null" json:"columns"` // JSON array of column names
	Unique    bool           `json:"unique"`
	SizeBytes int64          `json:"size_bytes"`
	Status    string         `gorm:"default:active" json:"status"` // active, inactive, maintenance
	Metadata  string         `gorm:"type:text" json:"metadata"`    // JSON string for additional data
}

// Backup represents a database backup.
type Backup struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DatabaseID  uint           `gorm:"not null;index" json:"database_id"`
	Database    Database       `gorm:"foreignKey:DatabaseID" json:"database"`
	Name        string         `gorm:"not null;index" json:"name"`
	Type        string         `gorm:"not null" json:"type"` // full, incremental, differential
	SizeBytes   int64          `json:"size_bytes"`
	Status      string         `gorm:"not null" json:"status"` // pending, in_progress, completed, failed
	FilePath    string         `json:"file_path"`
	Checksum    string         `json:"checksum"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Migration represents a database migration.
type Migration struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	DatabaseID   uint           `gorm:"not null;index" json:"database_id"`
	Database     Database       `gorm:"foreignKey:DatabaseID" json:"database"`
	Version      string         `gorm:"not null;index" json:"version"`
	Name         string         `gorm:"not null" json:"name"`
	Description  string         `json:"description"`
	SQL          string         `gorm:"type:text" json:"sql"`
	Status       string         `gorm:"not null" json:"status"` // pending, applied, failed, rolled_back
	AppliedAt    *time.Time     `json:"applied_at"`
	RolledBackAt *time.Time     `json:"rolled_back_at"`
	Error        string         `json:"error"`
	Metadata     string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// QueryResult represents the result of a SQL query.
type QueryResult struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	DatabaseID uint      `gorm:"not null;index" json:"database_id"`
	Query      string    `gorm:"type:text;not null" json:"query"`
	Rows       string    `gorm:"type:text" json:"rows"` // JSON array of rows
	RowCount   int       `json:"row_count"`
	Duration   int64     `json:"duration"` // in milliseconds
	Timestamp  time.Time `json:"timestamp"`
}

// TransactionOperation represents a database transaction operation.
type TransactionOperation struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	DatabaseID uint      `gorm:"not null;index" json:"database_id"`
	Type       string    `gorm:"not null" json:"type"` // insert, update, delete, select
	Table      string    `gorm:"not null" json:"table"`
	SQL        string    `gorm:"type:text;not null" json:"sql"`
	Parameters string    `gorm:"type:text" json:"parameters"` // JSON array of parameters
	Timestamp  time.Time `json:"timestamp"`
}

// DatabaseStats represents database statistics.
type DatabaseStats struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	DatabaseID  uint      `gorm:"not null;index" json:"database_id"`
	TableCount  int       `json:"table_count"`
	RowCount    int64     `json:"row_count"`
	SizeBytes   int64     `json:"size_bytes"`
	IndexCount  int       `json:"index_count"`
	LastBackup  time.Time `json:"last_backup"`
	LastUpdated time.Time `json:"last_updated"`
}

// TableStats represents table statistics.
type TableStats struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	TableID     uint      `gorm:"not null;index" json:"table_id"`
	RowCount    int64     `json:"row_count"`
	SizeBytes   int64     `json:"size_bytes"`
	IndexCount  int       `json:"index_count"`
	LastUpdated time.Time `json:"last_updated"`
}

// TableName returns the table name for Database.
func (Database) TableName() string {
	return "databases"
}

// TableName returns the table name for Table.
func (Table) TableName() string {
	return "tables"
}

// TableName returns the table name for Column.
func (Column) TableName() string {
	return "columns"
}

// TableName returns the table name for Index.
func (Index) TableName() string {
	return "indexes"
}

// TableName returns the table name for Backup.
func (Backup) TableName() string {
	return "backups"
}

// TableName returns the table name for Migration.
func (Migration) TableName() string {
	return "migrations"
}

// TableName returns the table name for QueryResult.
func (QueryResult) TableName() string {
	return "query_results"
}

// TableName returns the table name for TransactionOperation.
func (TransactionOperation) TableName() string {
	return "transaction_operations"
}

// TableName returns the table name for DatabaseStats.
func (DatabaseStats) TableName() string {
	return "database_stats"
}

// TableName returns the table name for TableStats.
func (TableStats) TableName() string {
	return "table_stats"
}

