-- Moveric MVP Schema

CREATE TABLE IF NOT EXISTS transfers (
    id              TEXT PRIMARY KEY,
    file_name       TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    chunk_size_mb   INT NOT NULL DEFAULT 64,
    total_chunks    INT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'PENDING',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chunks (
    id          TEXT PRIMARY KEY,   -- {transfer_id}_{part_index:05d}
    transfer_id TEXT NOT NULL REFERENCES transfers(id),
    part_index  INT NOT NULL,       -- 1-based
    size_bytes  BIGINT NOT NULL,
    is_last     BOOLEAN NOT NULL DEFAULT FALSE,
    checksum    TEXT NOT NULL,      -- SHA-256 hex
    minio_key   TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'PENDING',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chunks_transfer_id ON chunks(transfer_id);
CREATE INDEX IF NOT EXISTS idx_chunks_status ON chunks(transfer_id, status);
