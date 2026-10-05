"""Finite model-aware acquisition research. Not a production learning policy."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import random
import statistics
import sys

CASES = ["noise05", "noise20", "all_relevant", "null", "correlated"]
ARMS = ["random", "label_entropy", "full_information", "target_gini"]


def entropy(p):
    return -sum(x * math.log(x) for x in p if x > 0)


def classes(weights, case):
    out = [0.0] * (16 if case == "all_relevant" else 4)
    for h, w in enumerate(weights):
        out[h if case == "all_relevant" else h % 4] += w
    return out


def likelihood(case):
    noise = .2 if case == "noise20" else .05
    errors = [noise, noise, .01, .01, min(.4, 2*noise), .05, .5, noise]
    table = []
    for test in range(8):
        row = []
        for h in range(16):
            b = [(h >> i) & 1 for i in range(4)]
            values = b + [b[0] ^ b[1], b[2] ^ b[3], 0, b[0] ^ b[2]]
            e = .5 if case == "null" else errors[test]
            row.append(1-e if values[test] else e)
        table.append(row)
    return table


def update(weights, row, outcome):
    w = [p * (q if outcome else 1-q) for p, q in zip(weights, row)]
    total = sum(w)
    return [p/total for p in w]


def select(weights, table, case, arm, rng):
    if arm == "random":
        return rng.randrange(8)
    scores = []
    for row in table:
        p = sum(w*q for w, q in zip(weights, row))
        if arm == "label_entropy":
            score = entropy([p, 1-p])
        else:
            after = [update(weights, row, y) for y in [False, True]]
            if arm == "full_information":
                score = entropy(weights) - (1-p)*entropy(after[0]) - p*entropy(after[1])
            else:
                # E[sum class_mass^2] - sum prior_class_mass^2 is exactly
                # expected reduction in posterior target-class Gini impurity.
                square = lambda w: sum(x*x for x in classes(w, case))
                score = (1-p)*square(after[0]) + p*square(after[1]) - square(weights)
        scores.append(score)
    return max(range(8), key=lambda i: (round(scores[i], 14), -i))


def brier(p, target):
    return sum((v - (i == target))**2 for i, v in enumerate(p))


def episode(case, seed):
    h = random.Random(seed*10).randrange(16)
    rng = random.Random(seed*10+1)
    tape = [[rng.random() for _ in range(8)] for _ in range(16)]
    table = likelihood(case)
    target = h if case == "all_relevant" else h % 4
    out = {"seed": seed, "case": case, "truth": h, "arms": {}}
    for arm in ARMS:
        w = [1/16]*16
        rng = random.Random(seed*10+2)
        trace = []
        for t in range(16):
            p = classes(w, case)
            test = select(w, table, case, arm, rng)
            # Simulator-only oracle access occurs AFTER policy selection.
            y = tape[0 if case == "correlated" else t][test] < table[test][h]
            trace.append({"forecast": p, "test": test, "outcome": y})
            w = update(w, table[test], y)
        final = classes(w, case)
        chosen = max(range(len(final)), key=final.__getitem__)
        out["arms"][arm] = {"trace": trace, "curve_brier": statistics.mean(brier(v["forecast"], target) for v in trace),
                             "final_brier": brier(final, target), "correct": chosen == target,
                             "confident_wrong": max(final) >= .9 and chosen != target,
                             "final": final}
    return out


def summarize(records):
    summaries, comparisons = [], []
    passed = True
    for split in [0, 1]:
        for case in CASES:
            rows = [r for r in records if r["split"] == split and r["case"] == case]
            for arm in ARMS:
                summary = {"split": split, "case": case, "arm": arm, "n": len(rows)}
                for metric in ["curve_brier", "final_brier", "correct", "confident_wrong"]:
                    summary[metric] = statistics.mean(r["arms"][arm][metric] for r in rows)
                summaries.append(summary)
            for control in ARMS[:3]:
                gains = [r["arms"][control]["curve_brier"]-r["arms"]["target_gini"]["curve_brier"] for r in rows]
                mean = statistics.mean(gains)
                radius = 3.3*statistics.stdev(gains)/math.sqrt(len(gains))
                comparisons.append({"split": split, "case": case, "control": control, "gain": mean, "lower": mean-radius, "upper": mean+radius})
                if split == 1:
                    if case in ["noise05", "noise20"] and control in ARMS[:2]:
                        passed = passed and mean >= .02 and mean-radius > 0
                    if case in ["all_relevant", "null"] and control == "random":
                        passed = passed and mean >= -.01
    return {"summaries": summaries, "comparisons": comparisons, "matched_model_pass": passed}


def run():
    records = []
    for split in [0, 1]:
        for j, case in enumerate(CASES):
            for i in range(256):
                r = episode(case, 2026100201*1000000 + split*100000+j*1000+i)
                r["split"] = split
                records.append(r)
    return {"records": records, **summarize(records)}


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: experiment.py NEW.json.gz NEW-summary.json")
    paths = [Path(p) for p in sys.argv[1:]]
    if any(p.exists() for p in paths) or paths[0] == paths[1]:
        raise SystemExit("refusing to overwrite outputs")
    root = Path(__file__).parent
    hashes = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in [Path(__file__), root/"PROTOCOL.md"]}
    result = run()
    result["hashes"] = hashes
    with paths[0].open("xb") as f, gzip.GzipFile(fileobj=f, mode="wb", mtime=0) as z:
        z.write(json.dumps(result).encode())
    result.pop("records")
    with paths[1].open("x") as f:
        json.dump(result, f, indent=2)
    print("matched_model_pass:", result["matched_model_pass"])
