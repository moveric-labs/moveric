# Moveric

**Moveric — API-First Resumable File Transfer Engine**

Moveric is a lightweight, API-first file transfer engine designed to handle large file movement with **resumability**, **integrity**, and **reliability**.

## Why Moveric

Traditional file transfer systems:

- restart transfers on failure
- are complex and expensive
- lack developer-friendly APIs

Moveric solves this by:

- breaking files into chunks
- allowing resume from failure point
- ensuring end-to-end data integrity
- exposing simple API-based integration

## Core Features (MVP)

- Chunk-based file transfer
- Resume from last successful chunk
- SHA-256 checksum validation
- Offset-based file assembly
- Transfer tracking (DB-backed)

## Status

🚧 MVP in progress

## License

Apache 2.0