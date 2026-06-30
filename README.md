# Image Search Engine

Visual product search using CLIP embeddings and FAISS. Upload an image, get visually similar products ranked by cosine similarity.

## Architecture

```
Client → Go API (:8080) → Python AI (:8000) → CLIP → FAISS → Results
                ↓
        Metadata (product info)
```

**Two services:**
- **Go API** — HTTP gateway, metadata enrichment, request routing
- **Python AI** — CLIP ViT-B/32 embedding + FAISS IndexFlatIP search

## Components

| Component | Purpose |
|-----------|---------|
| `cmd/api` | Go HTTP server |
| `cmd/downloader` | Parallel image downloader (50 workers) |
| `cmd/indexer` | CSV → metadata.json generator |
| `internal/api` | Handlers + router |
| `internal/domain` | Models + metadata repository |
| `internal/python` | Python service HTTP client |
| `services/ai` | FastAPI app, CLIP model, FAISS search |
| `docker/` | Dockerfiles + compose |

## API

### Search
```bash
POST /api/search
Content-Type: multipart/form-data

Fields: file (image), k (results, default 500)

curl -X POST http://localhost:8080/api/search -F "file=@shoe.jpg" -F "k=500"
```
```json
{
  "results": [
    { "id": 15, "productName": "Nike Shoes", "category": "Shoes", "price": 120.00, "score": 0.91 }
  ]
}
```

### Health
```
GET /health → {"status": "ok"}
```

## Homepage UI

A minimal web interface built with Go templates for interactive image search.

**Access:** `http://localhost:8080/`

**Features:**
- Image upload form for easy searching
- Configurable result count (k parameter, 1-1000)
- Real-time results displayed in a responsive grid
- Product details: name, category, price, similarity score
- Product images displayed with each result
- Error handling with user-friendly messages

**Template:** `internal/api/templates/search.html`

## Setup

```bash
# 1. Download images
go run cmd/downloader/main.go

# 2. Build FAISS index
cd services/ai
python index_builder.py
cd ../..

# 3. Generate metadata
go run cmd/indexer/main.go

# 4. Start Python AI in one terminal
cd services/ai
source .venv/bin/activate
python -m uvicorn app:app --host 0.0.0.0 --port 8000

# 5. Start Go API from the repo root in another terminal
go run cmd/api/main.go
```
* Note : better run this way rather than docker compose because ai server may take time to be ready.
## Config

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Go API port |
| `PYTHON_SERVICE_URL` | `http://localhost:8000` | Python AI URL |
| `METADATA_PATH` | `index/metadata.json` | Metadata file path |
| `FAISS_INDEX_PATH` | auto-detects `index/faiss.index` | FAISS index path |
| `INDEX_WORKERS` | `4` | Parallel embedding threads |

## Latest Load Test

Results from k6 runs against the local Go API on different scenarios.

| Scenario | Requests | Throughput | p50 latency | p95 latency | Max latency | Error rate |
|----------|----------|------------|-------------|-------------|-------------|------------|
| Baseline | 22 | 0.72 req/s | 75ms | 86.3ms | 333.2ms | 0% |
| Concurrency | 2,085 | 20.71 req/s | 131.8ms | 365ms | 599.9ms | 0% |
| Spike | 1,147 | 29.8 req/s | 225ms | 439.6ms | 599.3ms | 0% |
| Soak | 13,075 | 21.73 req/s | 89ms | 313.1ms | 1.4s | 0% |
