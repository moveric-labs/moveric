package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"moveric/config"
	"moveric/chunk"
	"moveric/db"
	minioclient "moveric/minio"
	natsclient "moveric/nats"
	"moveric/models"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: source-agent <file-path>")
	}
	filePath := os.Args[1]

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	database, err := db.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	mc, err := minioclient.New(cfg)
	if err != nil {
		log.Fatalf("minio: %v", err)
	}

	nc, err := natsclient.New(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("nats: %v", err)
	}
	defer nc.Close()

	// Open file
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("open file: %v", err)
	}
	defer f.Close()

	stat, _ := f.Stat()
	fileSize := stat.Size()
	fileName := stat.Name()

	// Create transfer record
	transferID := uuid.New().String()
	chunks := chunk.Plan(transferID, fileSize, cfg.DefaultChunkSizeMB)

	transfer := &models.Transfer{
		ID:            transferID,
		FileName:      fileName,
		FileSizeBytes: fileSize,
		ChunkSizeMB:   cfg.DefaultChunkSizeMB,
		TotalChunks:   len(chunks),
		Status:        models.TransferPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := database.CreateTransfer(ctx, transfer); err != nil {
		log.Fatalf("create transfer: %v", err)
	}
	log.Printf("transfer %s created — %d chunks", transferID, len(chunks))

	if err := database.UpdateTransferStatus(ctx, transferID, models.TransferInProgress); err != nil {
		log.Fatalf("update transfer status: %v", err)
	}

	// Upload chunks
	for _, cm := range chunks {
		data, checksum, err := chunk.ReadChunk(f, cm.Offset, cm.Size)
		if err != nil {
			log.Fatalf("read chunk %d: %v", cm.PartIndex, err)
		}

		// Upload to MinIO
		if err := mc.UploadChunk(ctx, cm.Key, bytes.NewReader(data), cm.Size, checksum); err != nil {
			log.Fatalf("upload chunk %d: %v", cm.PartIndex, err)
		}

		// Persist chunk record
		chunkRecord := &models.Chunk{
			ID:         cm.Key,
			TransferID: transferID,
			PartIndex:  cm.PartIndex,
			SizeBytes:  cm.Size,
			IsLast:     cm.IsLast,
			Checksum:   checksum,
			MinioKey:   cm.Key,
			Status:     models.ChunkUploaded,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := database.CreateChunk(ctx, chunkRecord); err != nil {
			log.Fatalf("save chunk %d: %v", cm.PartIndex, err)
		}

		// Generate pre-signed URL (valid 1 hour)
		presignedURL, err := mc.PresignedGetURL(ctx, cm.Key, time.Hour)
		if err != nil {
			log.Fatalf("presign chunk %d: %v", cm.PartIndex, err)
		}

		// Publish NATS event
		event := models.ChunkReadyEvent{
			TransferID:   transferID,
			ChunkID:      cm.Key,
			PartIndex:    cm.PartIndex,
			TotalChunks:  len(chunks),
			IsLast:       cm.IsLast,
			PresignedURL: presignedURL,
			Checksum:     checksum,
		}
		if err := nc.Publish(natsclient.SubjectChunkReady, event); err != nil {
			log.Fatalf("publish chunk event %d: %v", cm.PartIndex, err)
		}

		fmt.Printf("✓ chunk %05d/%05d uploaded\n", cm.PartIndex, len(chunks))
	}

	log.Printf("transfer %s — all chunks uploaded", transferID)
}
