"""Score the UTC/local-time design trace, leaving confirmation unrun."""

import hashlib
import json
from pathlib import Path
import sys


ROOT = Path(__file__).parent
REPO = ROOT.parent.parent
SET = ROOT / "pr-time-v1"
DIGEST = "0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f"


def load(path):
    return json.loads(path.read_text())


def main(raw_path, summary_path):
    if summary_path.exists():
        raise ValueError("summary already exists")
    raw = load(raw_path)
    oracle = load(SET / "oracle.json")
    queries = load(SET / "queries.json")
    corpus = load(SET / "corpus.json")
    metadata = load(SET / "source-metadata.json")
    assert hashlib.sha256((SET / "source.geojson").read_bytes()).hexdigest() == metadata["source_sha256"]
    assert len(corpus) == 24 and len({f["fixture_id"] for f in corpus}) == 24
    design_cases = {q["case_id"] for q in queries if q["split"] == "design"}
    confirmation_cases = {q["case_id"] for q in queries if q["split"] == "confirmation"}
    assert len(design_cases) == 26 and len(confirmation_cases) == 24
    assert design_cases.isdisjoint(confirmation_cases)
    assert set(oracle) == design_cases | confirmation_cases
    for path, digest in raw["Hashes"].items():
        assert hashlib.sha256((REPO / path).read_bytes()).hexdigest() == digest, path
    assert raw["Digest"] == DIGEST and len(raw["Results"]) == 52

    by_arm = {}
    for arm in ("focus", "priority"):
        selected = [r for r in raw["Results"] if r["Arm"] == arm]
        assert len(selected) == 26 and {r["Case"] for r in selected} == design_cases
        by_arm[arm] = {r["Case"]: r for r in selected}

    rows, errors = [], []
    corpus_ids = {f["fixture_id"] for f in corpus}
    for case in sorted(design_cases):
        a, b = by_arm["focus"][case], by_arm["priority"][case]
        for arm, r in (("focus", a), ("priority", b)):
            frontier = r["Frontier"]
            if len(frontier) != 24 or len(set(frontier)) != 24 or set(frontier) != corpus_ids:
                errors.append(f"{arm}/{case}: frontier mismatch")
            if len(r["Before"]) != 24 or r["Before"] != r["After"]:
                errors.append(f"{arm}/{case}: numeric ranker changed candidates")
            if not r["JournalMatched"]:
                errors.append(f"{arm}/{case}: journal mismatch")
            packet = [c["Fixture"] for c in r["Candidates"]]
            if len(packet) != 10 or len(set(packet)) != 10 or not set(packet) <= corpus_ids:
                errors.append(f"{arm}/{case}: invalid pack")
            target = oracle[case]["target"]
            rows.append({
                "case": case, "split": "design", "wording": oracle[case]["wording"],
                "arm": arm, "target": target, "frontier_has_target": bool(target and target in frontier),
                "top1": bool(target and packet and packet[0] == target),
                "survived": bool(target and target in packet), "packet": packet,
            })
        if a["Laws"] != b["Laws"] or len(a["Laws"]) != 24:
            errors.append(f"{case}: full-frontier laws changed")
        if a["OriginalScores"] != b["OriginalScores"] or len(a["OriginalScores"]) != 24:
            errors.append(f"{case}: original numeric scores changed")

    counts = {}
    for arm in ("focus", "priority"):
        counts[arm] = {}
        for wording in ("utc", "local", "absent"):
            subset = [r for r in rows if r["arm"] == arm and r["wording"] == wording]
            counts[arm][wording] = {
                "cases": len(subset), "top1": sum(r["top1"] for r in subset),
                "survived": sum(r["survived"] for r in subset),
                "frontier_has_target": sum(r["frontier_has_target"] for r in subset),
            }
    priority_local = [r for r in rows if r["arm"] == "priority" and r["wording"] == "local"]
    local_misses = [r["case"] for r in priority_local if not r["top1"]]
    summary = {
        "raw_sha256": hashlib.sha256(raw_path.read_bytes()).hexdigest(),
        "oracle_sha256": hashlib.sha256((SET / "oracle.json").read_bytes()).hexdigest(),
        "scorer_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "counts": counts, "priority_local_misses": local_misses,
        "invariant_errors": errors,
        "headroom_qualifies": len(local_misses) >= 4
        and all(r["frontier_has_target"] for r in priority_local) and not errors,
        "rows": rows,
    }
    with summary_path.open("x") as handle:
        json.dump(summary, handle, indent=2)
        handle.write("\n")
    print(json.dumps({k: summary[k] for k in (
        "counts", "priority_local_misses", "invariant_errors", "headroom_qualifies")}, indent=2))


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: score_pr_time.py RAW.json NEW-SUMMARY.json")
    main(Path(sys.argv[1]), Path(sys.argv[2]))
