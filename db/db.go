package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"moveric/models"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (d *DB) Close() {
	d.pool.Close()
}

// --- Transfer ---

func (d *DB) CreateTransfer(ctx context.Context, t *models.Transfer) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO transfers (id, file_name, file_size_bytes, chunk_size_mb, total_chunks, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		t.ID, t.FileName, t.FileSizeBytes, t.ChunkSizeMB, t.TotalChunks, t.Status, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (d *DB) UpdateTransferStatus(ctx context.Context, id string, status models.TransferStatus) error {
	_, err := d.pool.Exec(ctx, `
		UPDATE transfers SET status=$1, updated_at=$2 WHERE id=$3`,
		status, time.Now(), id,
	)
	return err
}

func (d *DB) GetTransfer(ctx context.Context, id string) (*models.Transfer, error) {
	t := &models.Transfer{}
	err := d.pool.QueryRow(ctx, `
		SELECT id, file_name, file_size_bytes, chunk_size_mb, total_chunks, status, created_at, updated_at
		FROM transfers WHERE id=$1`, id,
	).Scan(&t.ID, &t.FileName, &t.FileSizeBytes, &t.ChunkSizeMB, &t.TotalChunks, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// --- Chunk ---

func (d *DB) CreateChunk(ctx context.Context, c *models.Chunk) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO chunks (id, transfer_id, part_index, size_bytes, is_last, checksum, minio_key, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		c.ID, c.TransferID, c.PartIndex, c.SizeBytes, c.IsLast, c.Checksum, c.MinioKey, c.Status, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (d *DB) UpdateChunkStatus(ctx context.Context, id string, status models.ChunkStatus) error {
	_, err := d.pool.Exec(ctx, `
		UPDATE chunks SET status=$1, updated_at=$2 WHERE id=$3`,
		status, time.Now(), id,
	)
	return err
}

func (d *DB) GetCompletedPartIndexes(ctx context.Context, transferID string) ([]int, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT part_index FROM chunks WHERE transfer_id=$1 AND status=$2`,
		transferID, models.ChunkDelivered,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []int
	for rows.Next() {
		var p int
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	return parts, nil
}
