import time
import asyncio
import logging
from fastapi import FastAPI, UploadFile, File, HTTPException

from model import embed_image, ImageEmbeddingError
from search import load_index, search

logger = logging.getLogger("app")

app = FastAPI()
load_index()


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/search")
async def image_search(file: UploadFile = File(...), k: int = 500):
    contents = await file.read()
    start = time.time()

    try:
        embedding = await asyncio.to_thread(embed_image, __import__("io").BytesIO(contents))
        matches = search(embedding, k)
    except ImageEmbeddingError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception:
        logger.exception("Search failed")
        raise HTTPException(status_code=500, detail="Internal server error")

    duration = time.time() - start
    return {"matches": matches, "count": len(matches), "took_s": duration}


@app.post("/index")
async def index_image(file: UploadFile = File(...)):
    from search import add_to_index

    contents = await file.read()
    start = time.time()

    try:
        embedding = await asyncio.to_thread(embed_image, __import__("io").BytesIO(contents))
        new_id = add_to_index(embedding)
    except ImageEmbeddingError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception:
        logger.exception("Index failed")
        raise HTTPException(status_code=500, detail="Internal server error")

    duration = time.time() - start
    logger.info(f"index: id={new_id} duration={duration:.3f}s")
    return {"id": new_id, "took_s": duration}
