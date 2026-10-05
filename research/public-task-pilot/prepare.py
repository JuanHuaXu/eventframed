"""Validate public facts and export corpus/query/oracle as separate artifacts."""
import datetime
import hashlib
import json
from pathlib import Path
import sys

ROOT = Path(__file__).parent


def prepare():
    facts = json.loads((ROOT / "facts.json").read_text())
    assert len(facts) == 13 and len({v["id"] for v in facts}) == 13
    corpus, queries, oracle = [], [], {}
    for f in facts:
        datetime.date.fromisoformat(f["date"])
        assert f["source"].startswith("https://science.nasa.gov/")
        relation = "from" if f["event"] == "launch" else "with" if f["event"] == "last radio contact" else "near"
        text = f'{f["mission"]}: {f["event"]} {relation} {f["target"]} occurred on {f["date"]}.'
        # Corpus is ordinary factual memory, not evaluation instructions.
        corpus.append({"fixture_id": f["id"], "text": text, "source": f["source"]})
        question = f'According to the retained mission records, what is the date of {f["mission"]}\'s {f["event"]} {relation} {f["target"]}?'
        assert f["date"] not in question
        queries.append({"case_id": f["id"], "split": f["split"], "question": question})
        oracle[f["id"]] = {"answer": f["date"], "support": [f["id"]], "cluster": f["cluster"]}
    for key, split, q, answer in [
        ("m10-ambiguous", "design", "What is the date of the Mariner 10 Mercury flyby?", "NEEDS_CLARIFICATION"),
        ("m10-absent", "design", "What date is recorded for Mariner 10's fourth Mercury flyby?", "UNKNOWN"),
        ("v2-absent", "confirmation", "What date is recorded for Voyager 2's Mercury flyby?", "UNKNOWN"),
    ]:
        queries.append({"case_id": key, "split": split, "question": q})
        oracle[key] = {"answer": answer, "support": [], "cluster": "mariner10" if split == "design" else "voyager2"}
    return {"corpus": corpus, "queries": queries, "oracle": oracle}


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: prepare.py NEW_OUTPUT_DIR")
    out = Path(sys.argv[1])
    out.mkdir(exist_ok=False)
    parts = prepare()
    hashes = {}
    for name, rows in parts.items():
        data = json.dumps(rows, indent=2).encode()
        (out / f"{name}.json").write_bytes(data)
        hashes[f"{name}.json"] = hashlib.sha256(data).hexdigest()
    (out / "manifest.json").write_text(json.dumps({"hashes": hashes, "facts": 13, "tasks": 16,
        "design_tasks": 10, "confirmation_tasks": 6, "confirmation_clusters": 1, "model_runs": 0}, indent=2))
    print("Validated and exported13 public facts and16 tasks; model runs=0")
