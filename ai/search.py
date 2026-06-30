import os
from pathlib import Path
import faiss
import numpy as np
import threading

INDEX = None
INDEX_PATH = None
_lock = threading.Lock()


def default_index_path():
    candidates = [
        Path("/app/index/faiss.index"),
        Path(__file__).resolve().parents[1] / "index" / "faiss.index",
    ]
    for candidate in candidates:
        if candidate.exists():
            return str(candidate)
    return str(candidates[-1])


def load_index(path=None):
    global INDEX, INDEX_PATH
    if path is None:
        path = os.environ.get("FAISS_INDEX_PATH", default_index_path())
    INDEX_PATH = path
    INDEX = faiss.read_index(path)


def search(vector, k=500):
    if INDEX is None:
        raise RuntimeError("Index not loaded")
    vector = vector.reshape(1, -1).astype(np.float32)
    distances, ids = INDEX.search(vector, k)
    return [
        {"id": int(idx), "score": float(score)}
        for idx, score in zip(ids[0], distances[0])
        if idx != -1
    ]


def add_to_index(vector):
    global INDEX
    if INDEX is None:
        raise RuntimeError("Index not loaded")
    vector = vector.reshape(1, -1).astype(np.float32)
    with _lock:
        new_id = INDEX.ntotal
        INDEX.add(vector)
        faiss.write_index(INDEX, INDEX_PATH)
    return new_id
