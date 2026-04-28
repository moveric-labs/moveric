# Moveric

**API-first resumable file transfer engine for enterprise environments.**

Moveric is an open-source Managed File Transfer (MFT) engine built for reliability, not convenience. It chunks large files, stages them through object storage, validates checksums end-to-end, and resumes from the exact point of failure — no restarts, no guesswork.

---

## Why Moveric

Traditional MFT tools are either too expensive, too brittle, or too opaque. Moveric is built differently:

- **Resumable by default** — transfers pick up from the last successful chunk, not from zero
- **Integrity guaranteed** — SHA-256 checksum validation on every chunk and final assembly
- **API-first** — every transfer operation is driven through a REST control plane
- **Observable** — transfer and chunk state is queryable at any point
- **Open core** — the transfer engine is Apache 2.0; enterprise features (multi-tenancy, HA, SSO) are commercial

---

## Features

- Chunk-based file splitting (configurable chunk size)
- Resumable transfers — survive network failures and process restarts
- SHA-256 checksum validation per chunk and assembled file
- MinIO as transit staging (chunks deleted post-assembly)
- NATS JetStream for event-driven chunk coordination
- PostgreSQL-backed transfer and chunk state tracking
- Source and destination agents as separate binaries
- REST API control plane (in progress)

---

## Architecture

**Stack:** Go · PostgreSQL · NATS JetStream · MinIO · Podman Compose

```mermaid
flowchart LR
    A[Source Agent] -->|splits into chunks| B[MinIO\nStaging]
    A -->|records transfer + chunk state| C[(PostgreSQL)]
    B -->|pre-signed URL via NATS event| D[NATS JetStream]
    D -->|chunk-ready event| E[Dest Agent]
    E -->|pulls chunk| B
    E -->|SHA-256 validate| E
    E -->|assembles file| F[Destination\nFile System]
    E -->|updates chunk state| C
    G[REST API / Control Plane] -->|reads/writes| C
```

---

## Roadmap

| Phase | Scope |
|-------|-------|
| ✅ MVP | Chunk-based transfer, SHA-256 validation, source/dest agents, Postgres state tracking |
| 🔄 Alpha | REST API control plane, resumable transfers, dashboard (React + Vite) |
| 🔜 Beta | Multi-tenancy, agent registration, SSE real-time updates, HA |
| 🔒 Enterprise | SSO, per-tenant quotas, rate limiting, audit logs |

---

## Quick Start (Dev)

### Prerequisites

- Go 1.21+
- Podman + podman-compose
- PostgreSQL client (`psql`)

### 1. Start infrastructure

```bash
cd moveric
podman-compose up -d
```

### 2. Apply schema

```bash
psql "postgres://moveric:moveric@127.0.0.1:5432/moveric_dev" -f schema.sql
```

### 3. Build binaries

```bash
# Source agent
go build -o bin/source-agent .

# Destination agent
go build -o bin/dest-agent ./dest/
```

### 4. Create a test file

```bash
# 2 GB test file
dd if=/dev/urandom of=/tmp/testfile bs=1M count=2000
```

### 5. Run source agent

```bash
./bin/source-agent /tmp/testfile
```

Expected output:

```
transfer created — 1 chunks
✓ chunk 00001/00001 uploaded
transfer — all chunks uploaded
```

### 6. Run destination agent

**Terminal 1 — start dest agent:**

```bash
./bin/dest-agent
```

**Terminal 2 — trigger transfer:**

```bash
./bin/source-agent /tmp/testfile
```

---

## Monitoring

### MinIO Console

```
http://127.0.0.1:9001
Login: moveric / moveric123
```

### PostgreSQL — transfer state

```bash
psql "postgres://moveric:moveric@127.0.0.1:5432/moveric_dev" \
  -c "SELECT id, status, total_chunks FROM transfers;"
```

### PostgreSQL — chunk state (live watch)

```bash
watch 'psql "postgres://moveric:moveric@127.0.0.1:5432/moveric_dev" \
  -c "SELECT id, part_index, status, checksum FROM chunks;"'
```

---

## Contributing

Contributions are welcome. The project is in active MVP development — the best way to contribute right now is to:

1. Try the quick start and report issues
2. Review open issues before submitting PRs
3. Keep PRs focused and small

Please open an issue before starting significant work.

---

## License

[Apache 2.0](LICENSE)
