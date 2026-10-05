"""Score retained public RFC rankings after the runner has saved raw output."""

import hashlib
import json
import math
from pathlib import Path
import sys


ROOT = Path(__file__).parent
REPO = ROOT.parent.parent
SET = ROOT / "rfc-transfer-v1"


def load(path):
    return json.loads(path.read_text())


def main(raw_path, summary_path):
    if summary_path.exists():
        raise ValueError("summary already exists")
    raw = load(raw_path)
    oracle = load(SET / "oracle.json")
    queries = load(SET / "queries.json")
    corpus = load(SET / "corpus.json")
    assert len(corpus) == 28 and len(queries) == 32 and len(oracle) == 32
    assert len({f["fixture_id"] for f in corpus}) == 28
    assert len({q["case_id"] for q in queries}) == 32
    for path, digest in raw["Hashes"].items():
        assert hashlib.sha256((REPO / path).read_bytes()).hexdigest() == digest, path
    assert raw["Digest"] == "0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f"

    cases = {q["case_id"] for q in queries}
    by_arm = {}
    for arm in ("focus", "priority"):
        selected = [r for r in raw["Results"] if r["Arm"] == arm]
        assert len(selected) == 32 and {r["Case"] for r in selected} == cases
        by_arm[arm] = {r["Case"]: r for r in selected}
    assert len(raw["Results"]) == 64

    rows = []
    invariant_errors = []
    for case in sorted(cases):
        a, b = by_arm["focus"][case], by_arm["priority"][case]
        for arm, r in (("focus", a), ("priority", b)):
            if r["Frontier"] != 28 or len(r["Before"]) != 28 or len(r["After"]) != 28:
                invariant_errors.append(f"{arm}/{case}: frontier changed")
            if r["Before"] != r["After"]:
                invariant_errors.append(f"{arm}/{case}: numeric ranker changed candidates")
            if not r["JournalMatched"]:
                invariant_errors.append(f"{arm}/{case}: packet/journal explanation mismatch")
            fixtures = [c["Fixture"] for c in r["Candidates"]]
            if len(fixtures) != len(set(fixtures)) or not 1 <= len(fixtures) <= 10:
                invariant_errors.append(f"{arm}/{case}: invalid packed identity")
            if any(f not in {x["fixture_id"] for x in corpus} for f in fixtures):
                invariant_errors.append(f"{arm}/{case}: foreign packed identity")
        if a["Laws"] != b["Laws"] or len(a["Laws"]) != 28:
            invariant_errors.append(f"{case}: full-frontier forecast laws changed")
        if a["OriginalScores"] != b["OriginalScores"] or len(a["OriginalScores"]) != 28:
            invariant_errors.append(f"{case}: original numeric candidate scores changed")
        support = set(oracle[case]["support"])
        for arm, r in (("focus", a), ("priority", b)):
            packed = [c["Fixture"] for c in r["Candidates"]]
            score = r["Candidates"][0]["Score"] if packed else None
            if score is not None and not math.isfinite(score):
                invariant_errors.append(f"{arm}/{case}: nonfinite top score")
            rows.append({
                "case": case, "wording": oracle[case]["wording"], "cluster": oracle[case]["cluster"],
                "arm": arm, "top1": bool(packed and packed[0] in support),
                "survived": any(f in support for f in packed), "packed": packed,
                "recall_ns": r["RecallNS"], "explanation_present": bool(r["Explanation"]),
                "packet_status": r["PacketStatus"],
            })
    counts = {}
    for arm in ("focus", "priority"):
        counts[arm] = {}
        for wording in ("literal", "paraphrase", "absent"):
            subset = [r for r in rows if r["arm"] == arm and r["wording"] == wording]
            counts[arm][wording] = {
                "cases": len(subset), "top1": sum(r["top1"] for r in subset),
                "survived": sum(r["survived"] for r in subset),
                "packed": sum(bool(r["packed"]) for r in subset),
            }
    nonharm = all(
        counts["priority"][w][metric] >= counts["focus"][w][metric]
        for w in ("literal", "paraphrase") for metric in ("top1", "survived")
    )
    gain = sum(counts["priority"][w]["top1"] - counts["focus"][w]["top1"] for w in ("literal", "paraphrase"))
    summary = {
        "raw_sha256": hashlib.sha256(raw_path.read_bytes()).hexdigest(),
        "oracle_sha256": hashlib.sha256((SET / "oracle.json").read_bytes()).hexdigest(),
        "scorer_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        "counts": counts, "top1_gain": gain, "invariant_errors": invariant_errors,
        "finite_transfer_pass": nonharm and gain >= 1 and not invariant_errors,
        "rows": rows,
    }
    with summary_path.open("x") as handle:
        json.dump(summary, handle, indent=2)
        handle.write("\n")
    print(json.dumps({k: summary[k] for k in ("counts", "top1_gain", "invariant_errors", "finite_transfer_pass")}, indent=2))


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: score_rfc_transfer.py RAW.json NEW-SUMMARY.json")
    main(Path(sys.argv[1]), Path(sys.argv[2]))
