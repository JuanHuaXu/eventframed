"""Score the saved held-back USGS confirmation against its frozen gate."""

import hashlib
import json
import math
from pathlib import Path
import sys


ROOT = Path(__file__).parent
REPO = ROOT.parent.parent
SET = ROOT / "usgs-headroom-v1"
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
    assert raw["Digest"] == DIGEST and len(raw["Results"]) == 48

    by_arm = {}
    for arm in ("focus", "priority"):
        selected = [r for r in raw["Results"] if r["Arm"] == arm]
        assert len(selected) == 24 and {r["Case"] for r in selected} == confirmation_cases
        by_arm[arm] = {r["Case"]: r for r in selected}

    rows, errors = [], []
    corpus_ids = {f["fixture_id"] for f in corpus}
    for case in sorted(confirmation_cases):
        a, b = by_arm["focus"][case], by_arm["priority"][case]
        for arm, r in (("focus", a), ("priority", b)):
            frontier = r["Frontier"]
            if len(frontier) != 24 or len(set(frontier)) != 24 or set(frontier) != corpus_ids:
                errors.append(f"{arm}/{case}: frontier mismatch")
            if len(r["Before"]) != 24 or r["Before"] != r["After"]:
                errors.append(f"{arm}/{case}: numeric ranker changed candidates")
            if not r["JournalMatched"]:
                errors.append(f"{arm}/{case}: journal mismatch")
            if r["RecallNS"] <= 0:
                errors.append(f"{arm}/{case}: invalid Recall duration")
            packet = [c["Fixture"] for c in r["Candidates"]]
            if len(packet) != 10 or len(set(packet)) != 10 or not set(packet) <= corpus_ids:
                errors.append(f"{arm}/{case}: invalid pack")
            target = oracle[case]["target"]
            rows.append({
                "case": case, "split": "confirmation", "wording": oracle[case]["wording"],
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
        for wording in ("literal", "paraphrase"):
            subset = [r for r in rows if r["arm"] == arm and r["wording"] == wording]
            counts[arm][wording] = {
                "cases": len(subset), "top1": sum(r["top1"] for r in subset),
                "survived": sum(r["survived"] for r in subset),
                "frontier_has_target": sum(r["frontier_has_target"] for r in subset),
            }
    outcomes = {arm: {r["case"]: r for r in rows if r["arm"] == arm} for arm in counts}
    gains = [case for case in sorted(confirmation_cases)
             if not outcomes["focus"][case]["top1"] and outcomes["priority"][case]["top1"]]
    top1_losses = [case for case in sorted(confirmation_cases)
                   if outcomes["focus"][case]["top1"] and not outcomes["priority"][case]["top1"]]
    support_losses = [case for case in sorted(confirmation_cases)
                      if outcomes["focus"][case]["survived"] and not outcomes["priority"][case]["survived"]]
    p95_ns = {}
    for arm in ("focus", "priority"):
        durations = sorted(r["RecallNS"] for r in by_arm[arm].values())
        p95_ns[arm] = durations[math.ceil(.95 * len(durations)) - 1]
    gain_each_wording = all(any(oracle[case]["wording"] == wording for case in gains)
                            for wording in ("literal", "paraphrase"))
    survival_nonharm = all(counts["priority"][w]["survived"] >= counts["focus"][w]["survived"]
                           for w in ("literal", "paraphrase"))
    summary = {
        "raw_sha256": hashlib.sha256(raw_path.read_bytes()).hexdigest(),
        "oracle_sha256": hashlib.sha256((SET / "oracle.json").read_bytes()).hexdigest(),
        "scorer_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "counts": counts, "top1_gains": gains, "top1_losses": top1_losses,
        "support_losses": support_losses, "isolated_recall_p95_ns": p95_ns,
        "invariant_errors": errors,
        "finite_confirmation_pass": len(gains) >= 8 and gain_each_wording
        and not top1_losses and not support_losses and survival_nonharm
        and not errors and all(v < 100_000_000 for v in p95_ns.values()),
        "rows": rows,
    }
    with summary_path.open("x") as handle:
        json.dump(summary, handle, indent=2)
        handle.write("\n")
    print(json.dumps({k: summary[k] for k in (
        "counts", "top1_gains", "top1_losses", "support_losses",
        "isolated_recall_p95_ns", "invariant_errors", "finite_confirmation_pass")}, indent=2))


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: score_usgs_confirmation.py RAW.json NEW-SUMMARY.json")
    main(Path(sys.argv[1]), Path(sys.argv[2]))
