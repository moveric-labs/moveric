package models

import "time"

type TransferStatus string
type ChunkStatus string

const (
	TransferPending    TransferStatus = "PENDING"
	TransferInProgress TransferStatus = "IN_PROGRESS"
	TransferCompleted  TransferStatus = "COMPLETED"
	TransferFailed     TransferStatus = "FAILED"

	ChunkPending   ChunkStatus = "PENDING"
	ChunkUploaded  ChunkStatus = "UPLOADED"
	ChunkDelivered ChunkStatus = "DELIVERED"
	ChunkFailed    ChunkStatus = "FAILED"
)

type Transfer struct {
	ID            string         `json:"id"`
	FileName      string         `json:"file_name"`
	FileSizeBytes int64          `json:"file_size_bytes"`
	ChunkSizeMB   int            `json:"chunk_size_mb"`
	TotalChunks   int            `json:"total_chunks"`
	Status        TransferStatus `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type Chunk struct {
	ID          string      `json:"id"`           // {transfer_id}_{part_index:05d}
	TransferID  string      `json:"transfer_id"`
	PartIndex   int         `json:"part_index"`   // 1-based
	SizeBytes   int64       `json:"size_bytes"`
	IsLast      bool        `json:"is_last"`
	Checksum    string      `json:"checksum"`     // SHA-256
	MinioKey    string      `json:"minio_key"`    // object key in MinIO
	Status      ChunkStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// NATS event payloads

type ChunkReadyEvent struct {
	TransferID   string `json:"transfer_id"`
	ChunkID      string `json:"chunk_id"`
	PartIndex    int    `json:"part_index"`
	TotalChunks  int    `json:"total_chunks"`
	IsLast       bool   `json:"is_last"`
	PresignedURL string `json:"presigned_url"`
	Checksum     string `json:"checksum"`
}

type ChunkAckEvent struct {
	TransferID string      `json:"transfer_id"`
	ChunkID    string      `json:"chunk_id"`
	PartIndex  int         `json:"part_index"`
	Status     ChunkStatus `json:"status"`
}
