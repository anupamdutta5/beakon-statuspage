// Package models provides data models for the Event Store Service.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Stream represents an event stream.
type Stream struct {
	ID        string         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Type      string         `gorm:"not null;index" json:"type"` // aggregate type
	Version   int            `gorm:"default:0" json:"version"`
	Metadata  string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Event represents a domain event.
type Event struct {
	ID        string         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	StreamID  string         `gorm:"not null;index" json:"stream_id"`
	Stream    Stream         `gorm:"foreignKey:StreamID" json:"stream"`
	Type      string         `gorm:"not null;index" json:"type"` // event type
	Version   int            `gorm:"not null;index" json:"version"`
	Data      string         `gorm:"type:text;not null" json:"data"` // JSON event data
	Metadata  string         `gorm:"type:text" json:"metadata"`      // JSON string for additional data
	Timestamp time.Time      `gorm:"not null;index" json:"timestamp"`
}

// Snapshot represents a stream snapshot.
type Snapshot struct {
	ID        string         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	StreamID  string         `gorm:"not null;index" json:"stream_id"`
	Stream    Stream         `gorm:"foreignKey:StreamID" json:"stream"`
	Version   int            `gorm:"not null;index" json:"version"`
	Data      string         `gorm:"type:text;not null" json:"data"` // JSON snapshot data
	Metadata  string         `gorm:"type:text" json:"metadata"`      // JSON string for additional data
	Timestamp time.Time      `gorm:"not null;index" json:"timestamp"`
}

// Projection represents a read model projection.
type Projection struct {
	ID          string         `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"not null;uniqueIndex" json:"name"`
	Description string         `json:"description"`
	Query       string         `gorm:"type:text;not null" json:"query"` // projection query
	Status      string         `gorm:"default:stopped" json:"status"`   // running, stopped, error
	LastEvent   string         `gorm:"index" json:"last_event"`         // last processed event ID
	LastUpdated time.Time      `json:"last_updated"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// Subscription represents an event subscription.
type Subscription struct {
	ID          string         `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"not null;uniqueIndex" json:"name"`
	StreamID    string         `gorm:"index" json:"stream_id"`       // optional: subscribe to specific stream
	EventTypes  string         `gorm:"type:text" json:"event_types"` // JSON array of event types
	Endpoint    string         `gorm:"not null" json:"endpoint"`     // webhook endpoint
	Status      string         `gorm:"default:active" json:"status"` // active, inactive, error
	LastEvent   string         `gorm:"index" json:"last_event"`      // last processed event ID
	LastUpdated time.Time      `json:"last_updated"`
	Error       string         `json:"error"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string for additional data
}

// EventStoreStats represents event store statistics.
type EventStoreStats struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	TotalStreams     int       `json:"total_streams"`
	TotalEvents      int64     `json:"total_events"`
	TotalProjections int       `json:"total_projections"`
	TotalSnapshots   int       `json:"total_snapshots"`
	LastUpdated      time.Time `json:"last_updated"`
}

// StreamStats represents stream statistics.
type StreamStats struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	StreamID      string    `gorm:"not null;index" json:"stream_id"`
	EventCount    int64     `json:"event_count"`
	SnapshotCount int       `json:"snapshot_count"`
	LastEventAt   time.Time `json:"last_event_at"`
	LastUpdated   time.Time `json:"last_updated"`
}

// TableName returns the table name for Stream.
func (Stream) TableName() string {
	return "streams"
}

// TableName returns the table name for Event.
func (Event) TableName() string {
	return "events"
}

// TableName returns the table name for Snapshot.
func (Snapshot) TableName() string {
	return "snapshots"
}

// TableName returns the table name for Projection.
func (Projection) TableName() string {
	return "projections"
}

// TableName returns the table name for Subscription.
func (Subscription) TableName() string {
	return "subscriptions"
}

// TableName returns the table name for EventStoreStats.
func (EventStoreStats) TableName() string {
	return "event_store_stats"
}

// TableName returns the table name for StreamStats.
func (StreamStats) TableName() string {
	return "stream_stats"
}

