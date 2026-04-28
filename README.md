podman-compose up -d

cd /apps/moveric

# 1. Apply schema
psql "postgres://moveric:moveric@127.0.0.1:5432/moveric_dev" -f schema.sql

# 2. Start infra (if not already running)
podman-compose up -d

# 3. Build binaries
go build -o bin/source-agent .

# 4. Create a test file
dd if=/dev/urandom of=/tmp/testfile bs=1M count=2000
dd if=/dev/urandom of=/apps/testfie_2gb bs=1M count=2000

# 5. Run source agent
./bin/source-agent /tmp/testfile
```

Expected output:
```
transfer <uuid> created — 1 chunks
✓ chunk 00001/00001 uploaded
transfer <uuid> — all chunks uploaded

Build:
go build -o bin/dest-agent ./dest/

# Terminal 1 — start dest agent
./bin/dest-agent

# Terminal 2 — resend the file to trigger events
./bin/source-agent /tmp/testfile

Monitor:
MINIO:
    http://127.0.0.1:9001 in browser
    Login: moveric / moveric123

Postgres:
psql "postgres://moveric:moveric@127.0.0.1:5432/moveric_dev" -c "SELECT id, status, total_chunks FROM transfers;"

watch 'psql "postgres://moveric:moveric@127.0.0.1:5432/moveric_dev" -c "SELECT id, part_index, status, checksum FROM chunks;"'

