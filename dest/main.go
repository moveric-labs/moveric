package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"moveric/db"
	"moveric/models"
	natsclient "moveric/nats"
	minioclient "moveric/minio"
	"moveric/config"
)

const deliveryDir = "/tmp/moveric-delivered"

func main() {
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

	os.MkdirAll(deliveryDir, 0755)
	log.Println("dest-agent listening for chunks...")

	// track received chunks per transfer: transferID -> count
	received := make(map[string]int)

	err = nc.Subscribe(natsclient.SubjectChunkReady, "dest-agent", func(data []byte) error {
		var event models.ChunkReadyEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}

		log.Printf("← chunk %05d/%05d transfer=%s", event.PartIndex, event.TotalChunks, event.TransferID)

		// Pull chunk via pre-signed URL
		chunkData, err := pullChunk(event.PresignedURL)
		if err != nil {
			return fmt.Errorf("pull chunk %d: %w", event.PartIndex, err)
		}

		// Verify checksum
		h := sha256.Sum256(chunkData)
		if got := hex.EncodeToString(h[:]); got != event.Checksum {
			return fmt.Errorf("checksum mismatch chunk %d: want %s got %s", event.PartIndex, event.Checksum, got)
		}

		// Write part file
		partPath := filepath.Join(deliveryDir, fmt.Sprintf("%s_%05d.part", event.TransferID, event.PartIndex))
		if err := os.WriteFile(partPath, chunkData, 0644); err != nil {
			return fmt.Errorf("write part: %w", err)
		}

		// Update DB + delete from MinIO
		database.UpdateChunkStatus(ctx, event.ChunkID, models.ChunkDelivered)
		mc.DeleteChunk(ctx, event.ChunkID)

		received[event.TransferID]++

		// Assemble when all chunks received
		if received[event.TransferID] == event.TotalChunks {
			if err := assemble(event.TransferID, event.TotalChunks); err != nil {
				return fmt.Errorf("assemble: %w", err)
			}
			database.UpdateTransferStatus(ctx, event.TransferID, models.TransferCompleted)
			delete(received, event.TransferID)
			log.Printf("✓ transfer %s assembled → %s/%s.assembled", event.TransferID, deliveryDir, event.TransferID)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("subscribe: %v", err)
	}

	select {}
}

func pullChunk(url string) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func assemble(transferID string, totalChunks int) error {
	outPath := filepath.Join(deliveryDir, transferID+".assembled")
	tmp := outPath + ".tmp"

	out, err := os.Create(tmp)
	if err != nil {
		return err
	}

	for i := 1; i <= totalChunks; i++ {
		partPath := filepath.Join(deliveryDir, fmt.Sprintf("%s_%05d.part", transferID, i))
		f, err := os.Open(partPath)
		if err != nil {
			out.Close()
			return err
		}
		_, err = io.Copy(out, f)
		f.Close()
		if err != nil {
			out.Close()
			return err
		}
		os.Remove(partPath)
	}
	out.Close()

	return os.Rename(tmp, outPath)
}
