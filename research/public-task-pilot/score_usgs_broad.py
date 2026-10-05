"""Score the broad USGS current-contract run after its raw trace exists."""

import hashlib
import json
import math
from pathlib import Path
import sys


ROOT = Path(__file__).parent
REPO = ROOT.parent.parent
SET = ROOT / "usgs-broad-confirmation-v1"
DIGEST = "0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f"


def load(path):
    return json.loads(path.read_text())


def main(raw_path, summary_path):
    if summary_path.exists():
        raise ValueError("summary already exists")
    raw = load(raw_path)
    metadata = load(SET / "source-metadata.json")
    corpus = load(SET / "corpus.json")
    queries = load(SET / "queries.json")
    assert hashlib.sha256((SET / "source.geojson").read_bytes()).hexdigest() == metadata["source_sha256"]
    assert (SET / "captured-at.txt").read_text().strip() == metadata["captured_at"]
    assert len(corpus) == 24 and len({f["fixture_id"] for f in corpus}) == 24
    assert len(queries) == 72 and len({q["case_id"] for q in queries}) == 72
    assert {q["split"] for q in queries} == {"design", "confirmation"}
    for path, digest in raw["Hashes"].items():
        assert hashlib.sha256((REPO / path).read_bytes()).hexdigest() == digest, path
    assert raw["Digest"] == DIGEST and len(raw["Results"]) == 144
    by_arm = {}
    cases = {q["case_id"]: q for q in queries}
    for arm in ("focus", "task-lexical"):
        selected = [r for r in raw["Results"] if r["Arm"] == arm]
        assert len(selected) == 72 and {r["Case"] for r in selected} == set(cases)
        by_arm[arm] = {r["Case"]: r for r in selected}
    # No answer labels or target IDs are read until all structural checks above.
    oracle = load(SET / "oracle.json")
    assert set(oracle) == set(cases)
    corpus_ids = {f["fixture_id"] for f in corpus}
    errors, rows = [], []
    for case in sorted(cases):
        focus, lexical = by_arm["focus"][case], by_arm["task-lexical"][case]
        for arm, result in (("focus", focus), ("task-lexical", lexical)):
            before, after = result["Before"], result["After"]
            before_ids = [c["ID"] for c in before]
            after_ids = [c["ID"] for c in after]
            frontier = result["Frontier"]
            packet = [c["Fixture"] for c in result["Candidates"]]
            if len(frontier) != 24 or set(frontier) != corpus_ids or len(set(frontier)) != 24:
                errors.append(f"{arm}/{case}: frontier mismatch")
            if len(before) != 24 or len(after) != 24 or set(before_ids) != set(after_ids) or len(set(after_ids)) != 24:
                errors.append(f"{arm}/{case}: ranker identity mismatch")
            if arm == "focus" and before != after:
                errors.append(f"{arm}/{case}: passthrough changed candidates")
            if arm == "task-lexical":
                for rank, candidate in enumerate(after):
                    expected = 1 - rank / (len(after) + 1)
                    if not math.isclose(candidate["Score"], expected, abs_tol=1e-12):
                        errors.append(f"{arm}/{case}: rank score mismatch at {rank}")
                        break
            if not result["JournalMatched"] or result["RecallNS"] <= 0:
                errors.append(f"{arm}/{case}: journal or duration mismatch")
            if len(packet) != 10 or len(set(packet)) != 10 or not set(packet) <= corpus_ids:
                errors.append(f"{arm}/{case}: invalid packed candidates")
            target = oracle[case]["target"]
            rows.append({
                "case": case, "split": oracle[case]["split"], "wording": oracle[case]["wording"],
                "arm": arm, "target": target, "top1": bool(packet and packet[0] == target),
                "survived": target in packet, "frontier_has_target": target in frontier,
                "packet": packet,
            })
        if focus["OriginalScores"] != lexical["OriginalScores"] or len(focus["OriginalScores"]) != 24:
            errors.append(f"{case}: pre-rank scores changed")
        if focus["Laws"] != lexical["Laws"] or len(focus["Laws"]) != 24:
            errors.append(f"{case}: full-frontier scored laws changed")
    outcomes = {arm: {r["case"]: r for r in rows if r["arm"] == arm} for arm in by_arm}
    gains = [case for case in sorted(cases) if not outcomes["focus"][case]["top1"] and outcomes["task-lexical"][case]["top1"]]
    top1_losses = [case for case in sorted(cases) if outcomes["focus"][case]["top1"] and not outcomes["task-lexical"][case]["top1"]]
    survival_losses = [case for case in sorted(cases) if outcomes["focus"][case]["survived"] and not outcomes["task-lexical"][case]["survived"]]
    counts = {}
    for arm in by_arm:
        counts[arm] = {}
        for split in ("design", "confirmation"):
            counts[arm][split] = {}
            for wording in ("literal", "paraphrase", "offset"):
                subset = [r for r in rows if r["arm"] == arm and r["split"] == split and r["wording"] == wording]
                counts[arm][split][wording] = {
                    "cases": len(subset), "top1": sum(r["top1"] for r in subset),
                    "survived": sum(r["survived"] for r in subset),
                    "frontier_has_target": sum(r["frontier_has_target"] for r in subset),
                }
    p95_ns = {}
    for arm in by_arm:
        durations = sorted(r["RecallNS"] for r in by_arm[arm].values())
        p95_ns[arm] = durations[math.ceil(.95 * len(durations)) - 1]
    gain_each_wording = all(any(oracle[case]["wording"] == wording for case in gains)
                            for wording in ("literal", "paraphrase", "offset"))
    survival_nonharm = all(
        sum(counts["task-lexical"][split][wording]["survived"] for split in ("design", "confirmation")) >=
        sum(counts["focus"][split][wording]["survived"] for split in ("design", "confirmation"))
        for wording in ("literal", "paraphrase", "offset")
    )
    summary = {
        "raw_sha256": hashlib.sha256(raw_path.read_bytes()).hexdigest(),
        "oracle_sha256": hashlib.sha256((SET / "oracle.json").read_bytes()).hexdigest(),
        "scorer_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "counts": counts, "top1_gains": gains, "top1_losses": top1_losses,
        "survival_losses": survival_losses,
        "event_clusters_with_gain": len({oracle[case]["target"] for case in gains}),
        "event_clusters_with_loss": len({oracle[case]["target"] for case in top1_losses}),
        "isolated_recall_p95_ns": p95_ns, "invariant_errors": errors,
        "finite_transfer_pass": len(gains) >= 8 and gain_each_wording and not top1_losses
        and not survival_losses and survival_nonharm and not errors
        and all(value < 100_000_000 for value in p95_ns.values()),
        "rows": rows,
    }
    with summary_path.open("x") as handle:
        json.dump(summary, handle, indent=2)
        handle.write("\n")
    print(json.dumps({key: summary[key] for key in (
        "counts", "top1_gains", "top1_losses", "survival_losses", "event_clusters_with_gain",
        "event_clusters_with_loss", "isolated_recall_p95_ns", "invariant_errors", "finite_transfer_pass",
    )}, indent=2))


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: score_usgs_broad.py RAW.json NEW-SUMMARY.json")
    main(Path(sys.argv[1]), Path(sys.argv[2]))
