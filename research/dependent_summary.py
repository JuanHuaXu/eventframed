"""Audit and summarize the frozen v9 raw artifact; never refit or select arms."""
import gzip
import hashlib
import json
import math
from collections import defaultdict
from pathlib import Path
import sys


def summarize(path):
    raw = json.loads(gzip.decompress(path.read_bytes()))
    for name, source in raw["Sources"].items():
        assert hashlib.sha256(source.encode()).hexdigest() == raw["Hashes"][name]
    records = raw["Records"]
    assert len(records) == 192
    groups = defaultdict(list)
    keys = set()
    for r in records:
        key = (r["Generator"], r["Scenario"], r["Split"], r["Fit"], r["Stream"])
        assert key not in keys
        keys.add(key)
        assert len(r["Ticks"]) == len(r["Inputs"]) == len(r["Views"]) == 512
        start = 128 if r["Scenario"] in ("shift128", "recurring") else 256
        for t, tick in enumerate(r["Ticks"]):
            for origin in tick.get("Delivered") or []:
                assert origin <= t
            for arm in range(4):
                p = tick["Predictions"][arm]["P"]
                assert math.isfinite(p) and 0 < p < 1
                view = r["Views"][t][arm]
                assert 0 <= view["cost"] <= 6
                for step in view["trace"]:
                    assert step["values"] == (r["Inputs"][t] & step["observed"])
        for field, begin in (("Full", 0), ("Post", start)):
            for arm in range(4):
                ticks = r["Ticks"][begin:]
                brier = sum((v["Predictions"][arm]["P"] - v["Outcome"]) ** 2 for v in ticks) / len(ticks)
                accuracy = sum((v["Predictions"][arm]["P"] >= .5) == v["Outcome"] for v in ticks) / len(ticks)
                assert abs(brier - r[field][arm]["Brier"]) < 1e-12
                assert abs(accuracy - r[field][arm]["Accuracy"]) < 1e-12
        groups[key[:3]].append(r)
    comparisons, failures = [], defaultdict(list)
    for (generator, scenario, split), rows in sorted(groups.items()):
        assert len(rows) == 8
        for arm, name in enumerate(("fixed", "replacement", "adaptive_retained", "static_retained")):
            if arm == 0:
                continue
            c = {"generator": generator, "scenario": scenario, "split": split, "arm": name, "streams": 8}
            for window in ("Full", "Post"):
                b = sum(r[window][0]["Brier"] for r in rows) / 8
                a = sum(r[window][arm]["Brier"] for r in rows) / 8
                c[window] = {"baseline_brier": b, "candidate_brier": a, "harm": a-b}
                if arm >= 2 and a-b > .01:
                    failures[name].append({"kind": "harm", **c, "window": window})
            if arm >= 2 and scenario == "shift128" and c["Post"]["harm"] > -.005:
                failures[name].append({"kind": "insufficient_shift_gain", **c})
            comparisons.append(c)
    return {"raw_sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
            "records": len(records), "frames": len(records)*512,
            "comparisons": comparisons,
            "verdicts": {name: {"passed": not failures[name], "failures": failures[name]}
                         for name in ("adaptive_retained", "static_retained")}}


if __name__ == "__main__":
    result = summarize(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as out:
        json.dump(result, out, indent=2)
        out.write("\n")
    print(json.dumps({k: {"passed": v["passed"], "failures": len(v["failures"])}
                      for k, v in result["verdicts"].items()}))
