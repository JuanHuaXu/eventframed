"""Summarize consumed-data alternate-path diagnostics under frozen criteria."""
import gzip
import hashlib
import json
import math
from collections import defaultdict
from pathlib import Path
import sys


def summarize(path):
    raw = json.loads(gzip.decompress(path.read_bytes()))
    groups = defaultdict(list)
    keys = set()
    assert len(raw["Records"]) == 192
    count = 0
    for r in raw["Records"]:
        key = (r["Generator"], r["Scenario"], r["Split"], r["Fit"], r["Stream"])
        assert key not in keys
        keys.add(key)
        metrics = {}
        for window in ("Full", "Post"):
            start = 0 if window == "Full" else 128 if r["Scenario"] in ("shift128", "recurring") else 256
            ticks = [v for v in r["Ticks"] if v["Step"] >= start]
            assert ticks
            metrics[window] = [sum((v["Predictions"][a] - v["Outcome"]) ** 2 for v in ticks) / len(ticks) for a in range(4)]
            metrics[window+"Accuracy"] = [sum((v["Predictions"][a]>=.5)==v["Outcome"] for v in ticks)/len(ticks) for a in range(4)]
            metrics[window+"Cost"] = [sum(v["Views"][a]["cost"] for v in ticks)/len(ticks) for a in range(4)]
        for tick in r["Ticks"]:
            assert tick["Audits"] >= 32
            assert all(v["cost"] <= 6 and v["observed"] <= 6 for v in tick["Views"])
            assert all(math.isfinite(p) and 0 < p < 1 for p in tick["Predictions"])
        count += len(r["Ticks"])
        groups[key[:3]].append(metrics)
    comparisons, failures = [], []
    for key, rows in sorted(groups.items()):
        assert len(rows) == 8
        for window in ("Full", "Post"):
            means = [sum(r[window][a] for r in rows)/8 for a in range(4)]
            c = dict(zip(("generator", "scenario", "split"), key))
            c.update(window=window, short=means[0], uniform_tree=means[1], empirical=means[2], oracle=means[3], empirical_harm=means[2]-means[0], oracle_harm=means[3]-means[0])
            c["accuracy"] = [sum(r[window+"Accuracy"][a] for r in rows)/8 for a in range(4)]
            c["cost"] = [sum(r[window+"Cost"][a] for r in rows)/8 for a in range(4)]
            comparisons.append(c)
            if c["empirical_harm"] > .01:
                failures.append({"reason": "harm", **c})
            if key[0] == "clustered" and key[1] == "shift128" and window == "Post" and c["empirical_harm"] > -.005:
                failures.append({"reason": "insufficient_gain", **c})
    return {"sha256": hashlib.sha256(path.read_bytes()).hexdigest(), "input_sha256": raw["InputSHA256"], "records": 192, "scored_frames": count,
            "comparisons": comparisons, "passed": not failures, "failures": failures}


if __name__ == "__main__":
    result = summarize(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as f:
        json.dump(result,f,indent=2)
        f.write("\n")
    print(json.dumps({"passed": result["passed"], "failures":len(result["failures"]), "scored_frames":result["scored_frames"]}))
