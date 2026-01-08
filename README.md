# AudioML – Local Model Training & Versioning (Go + Python)

## Overview

AudioML is a small backend-oriented ML system that demonstrates **how to train, version, and manage ML models locally** using a clean Go + Python workflow.

The goal is not to build a complex AI, but to show **solid engineering foundations**:

* clear separation of responsibilities
* reproducible local training
* explicit model versioning
* artifacts managed as first‑class citizens

This project is designed to be extended later (MinIO/S3, inference API, async jobs).
---

## Architecture (Simple View)

```
┌─────────┐        ┌─────────────┐        ┌─────────────┐
│  Go API │─────▶ |  Python ML    ─────▶ │  Artifacts  │
│         │        │  Training   │        │ (model.bin) │
└─────────┘        └─────────────┘        └─────────────┘
      │                                         │
      └───────────────▶ PostgreSQL ◀───────────┘
                     (model_versions)
```

**Go**

* Orchestrates training
* Manages model versions
* Stores metadata in PostgreSQL

**Python**

* Handles ML logic (training)
* Produces model artifacts

**PostgreSQL**

* Source of truth for model versions
* Links models to artifact paths

---

## Workflow

1. Go triggers a training command
2. Python script trains the model
3. Model artifact is saved locally (`artifacts/<uuid>/model.bin`)
4. Go inserts a new row in `model_versions`
5. Latest model can be retrieved by version

Each training run = **one immutable model version**.

---

## Database Table (Simplified)

```
model_versions
- name
- version
- artifact_path
- created_at
```

This keeps the system auditable and reproducible.

---

## Database Table (Simplified)

This demo shows the full flow: **dataset upload → training → model versioning → artifacts**.

### 1. Start the API

```bash
go run ./cmd/api/main.go
```

You should see:

```bash
API listening on :8080
```

---

### 2. Upload a Dataset

```bash
curl -X POST http://localhost:8080/datasets/upload \
  -F"dataset=demo2" \
  -F"files=@C:\test\call_test1.wav" \
  -F"files=@C:\test\Downloads\call_test2.wav"
```

Response:

```json
{"dataset":"local-audio/demo2","files":2}

```

Verify files are stored locally:

```bash
ls ./datasets/local-audio/
demo2/

```

---

### 3. Start Training

```bash
curl -X POST http://localhost:8080/training/start \
  -H"Content-Type: application/json" \
  -d'{"dataset":"local-audio/demo2","model":"emotion"}'

```

Response:

```json
{
"ID":"78465fa9-8a59-4eff-ada5-b6169a06abed",
"Status":"queued",
"DatasetSource":"local-audio/demo2",
"ModelName":"emotion",
"CreatedAt":"2026-01-05T16:33:48.2986936+01:00"
}

```

---

### 4. Watch Training Logs

```bash
Python raw output:
{
"metrics": {"accuracy": 0.91,"loss": 0.08},
"params": {"epochs": 10,"lr": 0.001},
"artifact_path":"artifacts\\78465fa9-8a59-4eff-ada5-b6169a06abed\\model.bin"
}

```

---

### 5. Verify Model Version in Database

```sql
SELECT name, version, artifact_path
FROM model_versions
ORDERBY created_atDESC;

```

Result:

```
 name    | version |                      artifact_path
---------+---------+----------------------------------------------------------
 emotion |       4 | artifacts\78465fa9-8a59-4eff-ada5-b6169a06abed\model.bin

```

---

### 6. Check Generated Artifacts

```bash
ls artifacts/
78465fa9-8a59-4eff-ada5-b6169a06abed/

```

---

### What This Demonstrates

- Dataset ingestion and storage
- Async training execution
- Python ↔ Go integration
- Model versioning in database
- Artifact persistence on disk

## Next Possible Steps

* Model inference pipeline
* HTTP API for predictions
* Async training with queues (NATS)
* Artifact storage via MinIO / S3
* Metrics & logging

