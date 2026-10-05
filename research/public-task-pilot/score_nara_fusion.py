"""Score the saved National Archives raw trace against the frozen oracle."""

import hashlib
import json
import math
from pathlib import Path
import sys


ROOT = Path(__file__).parent
REPO = ROOT.parent.parent
SET = ROOT / "nara-fusion-v1"
DIGEST = "0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f"


def load(path):
    return json.loads(path.read_text())


def expected_fusion(base, priority):
    first = priority[0] if priority else None
    return [first] + [x for x in base if x != first] if first in base else list(base)


def main(raw_path, summary_path):
    if summary_path.exists():
        raise ValueError("summary already exists")
    raw = load(raw_path)
    oracle = load(SET / "oracle.json")
    queries = load(SET / "queries.json")
    corpus = load(SET / "corpus.json")
    facts = load(SET / "facts.json")
    assert len(facts) == 15 and len(corpus) == 28 and len(queries) == 32 and len(oracle) == 32
    assert len({f["fixture_id"] for f in corpus}) == 28
    cases = {q["case_id"] for q in queries}
    assert len(cases) == 32 and cases == set(oracle) == set(raw["Fusions"])
    for f in facts:
        assert f["year"] not in f["paraphrase"]
    for path, digest in raw["Hashes"].items():
        assert hashlib.sha256((REPO / path).read_bytes()).hexdigest() == digest, path
    assert raw["Digest"] == DIGEST

    by_arm = {}
    for arm in ("focus", "priority"):
        selected = [r for r in raw["Results"] if r["Arm"] == arm]
        assert len(selected) == 32 and {r["Case"] for r in selected} == cases
        by_arm[arm] = {r["Case"]: r for r in selected}
    assert len(raw["Results"]) == 64

    rows = []
    invariant_errors = []
    corpus_ids = {f["fixture_id"] for f in corpus}
    for case in sorted(cases):
        a, b = by_arm["focus"][case], by_arm["priority"][case]
        packets = {}
        for arm, r in (("focus", a), ("priority", b)):
            if r["Frontier"] != 28 or len(r["Before"]) != 28 or len(r["After"]) != 28:
                invariant_errors.append(f"{arm}/{case}: frontier changed")
            if r["Before"] != r["After"]:
                invariant_errors.append(f"{arm}/{case}: numeric ranker changed candidates")
            if not r["JournalMatched"]:
                invariant_errors.append(f"{arm}/{case}: packet/journal explanation mismatch")
            packed = [c["Fixture"] for c in r["Candidates"]]
            if len(packed) != len(set(packed)) or not 1 <= len(packed) <= 10:
                invariant_errors.append(f"{arm}/{case}: invalid packet IDs")
            if any(f not in corpus_ids for f in packed):
                invariant_errors.append(f"{arm}/{case}: foreign packed ID")
            if any(not math.isfinite(c["Score"]) for c in r["Candidates"]):
                invariant_errors.append(f"{arm}/{case}: nonfinite rank score")
            packets[arm] = packed
        if a["Laws"] != b["Laws"] or len(a["Laws"]) != 28:
            invariant_errors.append(f"{case}: full-frontier laws changed")
        if a["OriginalScores"] != b["OriginalScores"] or len(a["OriginalScores"]) != 28:
            invariant_errors.append(f"{case}: original numeric scores changed")
        fused = raw["Fusions"][case]
        packets["fusion"] = fused
        if fused != expected_fusion(packets["focus"], packets["priority"]):
            invariant_errors.append(f"{case}: fusion rule mismatch")
        if len(fused) != len(set(fused)) or len(fused) > 10 or set(fused) != set(packets["focus"]):
            invariant_errors.append(f"{case}: baseline support set changed")
        target = oracle[case]["target"]
        for arm, packed in packets.items():
            rows.append({
                "case": case, "wording": oracle[case]["wording"], "target": target,
                "arm": arm, "top1": bool(target and packed and packed[0] == target),
                "survived": bool(target and target in packed), "packed": packed,
            })

    counts = {}
    for arm in ("focus", "priority", "fusion"):
        counts[arm] = {}
        for wording in ("literal", "paraphrase", "absent"):
            subset = [r for r in rows if r["arm"] == arm and r["wording"] == wording]
            counts[arm][wording] = {
                "cases": len(subset), "top1": sum(r["top1"] for r in subset),
                "survived": sum(r["survived"] for r in subset),
                "packed": sum(bool(r["packed"]) for r in subset),
            }
    outcomes = {arm: {r["case"]: r for r in rows if r["arm"] == arm} for arm in counts}
    losses = [case for case in sorted(cases) if oracle[case]["target"]
              and outcomes["focus"][case]["top1"] and not outcomes["fusion"][case]["top1"]]
    gains = [case for case in sorted(cases) if oracle[case]["target"]
             and not outcomes["focus"][case]["top1"] and outcomes["fusion"][case]["top1"]]
    priority_losses = [case for case in sorted(cases) if oracle[case]["target"]
                       and outcomes["focus"][case]["top1"] and not outcomes["priority"][case]["top1"]]
    summary = {
        "raw_sha256": hashlib.sha256(raw_path.read_bytes()).hexdigest(),
        "oracle_sha256": hashlib.sha256((SET / "oracle.json").read_bytes()).hexdigest(),
        "scorer_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "counts": counts, "fusion_top1_gains": gains, "fusion_top1_losses": losses,
        "priority_top1_losses": priority_losses, "invariant_errors": invariant_errors,
        "finite_fusion_pass": bool(gains) and not losses and not invariant_errors,
        "rows": rows,
    }
    with summary_path.open("x") as handle:
        json.dump(summary, handle, indent=2)
        handle.write("\n")
    print(json.dumps({k: summary[k] for k in (
        "counts", "fusion_top1_gains", "fusion_top1_losses", "priority_top1_losses",
        "invariant_errors", "finite_fusion_pass")}, indent=2))


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: score_nara_fusion.py RAW.json NEW-SUMMARY.json")
    main(Path(sys.argv[1]), Path(sys.argv[2]))
