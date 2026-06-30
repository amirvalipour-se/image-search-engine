import os
import time
import faiss
import numpy as np
import pandas as pd
import json
from concurrent.futures import ThreadPoolExecutor, as_completed

from model import embed_image, ImageEmbeddingError

CSV_PATH = "../data/amazon_products_all.csv"
IMAGES_DIR = "../images"
MAX_WORKERS = int(os.environ.get("INDEX_WORKERS", 4))

print("Loading CSV...")
df = pd.read_csv(CSV_PATH)
print(f"CSV loaded: {len(df)} rows")


def embed_one(i):
    image_path = os.path.join(IMAGES_DIR, f"{i}.jpg")
    if not os.path.exists(image_path):
        return i, None, "not_found"
    try:
        embedding = embed_image(image_path)
        return i, embedding, None
    except ImageEmbeddingError as e:
        return i, None, f"embedding_failed: {e}"
    except Exception as e:
        return i, None, f"unexpected_error: {e}"


print(f"Embedding with {MAX_WORKERS} workers...")
start = time.time()

results = {}
with ThreadPoolExecutor(max_workers=MAX_WORKERS) as pool:
    futures = {pool.submit(embed_one, i): i for i in range(len(df))}
    done_count = 0
    for future in as_completed(futures):
        i, embedding, error = future.result()
        done_count += 1
        if embedding is not None:
            results[i] = embedding
        if error and done_count % 100 == 0:
            print(f"  [{done_count}/{len(df)}] {error}")
        elif done_count % 100 == 0:
            print(f"  [{done_count}/{len(df)}] embedded")

elapsed = time.time() - start
print(f"Embedding done in {elapsed:.1f}s — {len(results)} success, {len(df) - len(results)} skipped")

sorted_ids = sorted(results.keys())
embeddings = [results[i].reshape(1, -1) for i in sorted_ids]
dim = embeddings[0].shape[1]

index = faiss.IndexFlatIP(dim)
index.add(np.vstack(embeddings))

print(f"FAISS index built: {index.ntotal} vectors")

os.makedirs("index", exist_ok=True)
faiss.write_index(index, "index/faiss.index")

with open("index/image_ids.json", "w") as f:
    json.dump(sorted_ids, f)

skipped_count = len(df) - len(results)
print(f"Done! Indexed {index.ntotal} images, skipped {skipped_count}")