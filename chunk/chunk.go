package chunk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const DefaultChunkSizeMB = 64

type ChunkMeta struct {
	PartIndex int    // 1-based
	Offset    int64
	Size      int64
	IsLast    bool
	Checksum  string // SHA-256 hex
	Key       string // MinIO object key: {transferID}_{partIndex:05d}
}

// Plan returns chunk metadata for a file without reading data.
func Plan(transferID string, fileSize int64, chunkSizeMB int) []ChunkMeta {
	chunkSize := int64(chunkSizeMB) * 1024 * 1024
	total := int((fileSize + chunkSize - 1) / chunkSize)
	if total == 0 {
		total = 1
	}

	chunks := make([]ChunkMeta, total)
	for i := 0; i < total; i++ {
		partIndex := i + 1
		offset := int64(i) * chunkSize
		size := chunkSize
		if offset+size > fileSize {
			size = fileSize - offset
		}
		chunks[i] = ChunkMeta{
			PartIndex: partIndex,
			Offset:    offset,
			Size:      size,
			IsLast:    partIndex == total,
			Key:       ChunkKey(transferID, partIndex),
		}
	}
	return chunks
}

// ChunkKey returns the MinIO object key for a chunk.
func ChunkKey(transferID string, partIndex int) string {
	return fmt.Sprintf("%s_%05d", transferID, partIndex)
}

// ReadChunk reads a chunk from a file and computes its SHA-256.
func ReadChunk(f *os.File, offset, size int64) ([]byte, string, error) {
	buf := make([]byte, size)
	_, err := f.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return nil, "", fmt.Errorf("read chunk at offset %d: %w", offset, err)
	}

	h := sha256.Sum256(buf)
	checksum := hex.EncodeToString(h[:])
	return buf, checksum, nil
}
