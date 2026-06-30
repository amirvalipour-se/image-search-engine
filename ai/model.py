import os
os.environ.setdefault("KMP_DUPLICATE_LIB_OK", "TRUE")
for _var, _val in (("OMP_NUM_THREADS", "1"), ("OPENBLAS_NUM_THREADS", "1"), ("MKL_NUM_THREADS", "1"), ("VECLIB_MAXIMUM_THREADS", "1"), ("NUMEXPR_NUM_THREADS", "1")):
    os.environ.setdefault(_var, _val)

import torch
import open_clip
from PIL import Image
import numpy as np


class ImageEmbeddingError(Exception):
    pass


def _choose_device():
    if hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
        return "mps"
    if torch.cuda.is_available():
        return "cuda"
    return "cpu"


device = _choose_device()
model, _, preprocess = open_clip.create_model_and_transforms(
    "ViT-B-32", pretrained="laion2b_s34b_b79k"
)
model = model.to(device).eval()

USE_FP16 = os.environ.get("MODEL_FP16", "0") in ("1", "true", "True")
if USE_FP16 and device != "cpu":
    try:
        model.half()
    except Exception:
        pass


@torch.no_grad()
def embed_image(path_or_file):
    try:
        image = Image.open(path_or_file).convert("RGB")
    except Exception as e:
        raise ImageEmbeddingError(f"Failed to load image: {e}") from e

    try:
        image = preprocess(image).unsqueeze(0).to(device)
    except Exception as e:
        raise ImageEmbeddingError(f"Failed to preprocess: {e}") from e

    if USE_FP16 and device != "cpu":
        try:
            image = image.half()
        except Exception:
            pass

    try:
        if device == "cuda":
            with torch.cuda.amp.autocast(enabled=True):
                embedding = model.encode_image(image)
        else:
            embedding = model.encode_image(image)
    except Exception as e:
        raise ImageEmbeddingError(f"Failed to encode: {e}") from e

    embedding = embedding / embedding.norm(dim=-1, keepdim=True)
    return embedding.squeeze().cpu().numpy().astype(np.float32)
