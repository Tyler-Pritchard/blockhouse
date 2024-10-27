package models

// Stream represents a data model for a streaming session, identified by a unique ID
// and carrying associated data as a string payload.
type Stream struct {
	ID   string `json:"id"`   // Unique identifier for the stream
	Data string `json:"data"` // Data payload associated with the stream
}
