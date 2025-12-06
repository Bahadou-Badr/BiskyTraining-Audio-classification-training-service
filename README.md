# Go-based Scalable Audio Classification Training Platform (MVP)

## Phase 0 — Setup & research (this repo)
This repository contains the Phase 0 scaffold: a Go backend starter, local dev stack (`postgres`, `nats`, `minio`), and a containerized Python trainer placeholder.

### Quick start (local)
1. Copy env:
   ```bash
   cp .env.example .env

```bash
docker-compose -f deploy/docker-compose.yml up --build
```