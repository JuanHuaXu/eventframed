"""Reconstruct frozen acquisition choices and compare the complete experiment."""
import gzip
import hashlib
import json
from pathlib import Path
import experiment

root = Path(__file__).resolve().parents[2]
with gzip.open(root/"docs/experiments/mmm-hypothesis-v1.json.gz", "rt") as f:
    actual = json.load(f)
for name, expected in actual.pop("hashes").items():
    assert hashlib.sha256((Path(__file__).parent/name).read_bytes()).hexdigest() == expected, name
assert len(actual["records"]) == 2560
assert len({r["seed"] for r in actual["records"]}) == 2560
assert actual == experiment.run(), "full replay differs"
print("Full replay, seeds, hashes, summaries and decisions verified")
