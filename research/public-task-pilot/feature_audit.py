"""Design-only representation diagnostic, not learned performance evidence."""
import json
import re
from pathlib import Path


def coverage_key(query, frame):
    # Retain the identity of every query term matched by the compressed frame.
    # No answer, record ID, or oracle is consulted to build this signature.
    q = set(re.findall(r"[^\W_]+", query.lower()))
    f = set(re.findall(r"[^\W_]+", frame.lower()))
    return tuple(sorted(q & f))


def main():
    root = Path(__file__).parent
    rows = json.loads((root / "design-preflight-v1.json").read_text())
    queries = {q["case_id"]: q for q in json.loads(
        (root / "prepared-v2/queries.json").read_text())}
    oracle = json.loads((root / "prepared-v2/oracle.json").read_text())
    totals = [0, 0, 0]
    for row in rows:
        query = queries[row["case"]]
        assert query["split"] == "design"
        keys = [coverage_key(query["question"], c["frame"])
                for c in row["candidates"]]
        support = oracle[row["case"]]["support"]
        if not support:
            continue
        for i, c in enumerate(row["candidates"]):
            if c["fixture"] not in support:
                continue
            collisions = sum(keys[j] == keys[i] for j, other in
                             enumerate(row["candidates"])
                             if other["fixture"] not in support)
            totals[0] += 1
            totals[1] += row["support_feature_collisions"] > 0
            totals[2] += collisions > 0
            print(row["case"], "nine_bit_collisions=",
                  row["support_feature_collisions"], "token_collisions=", collisions)
    print("supported queries, nine-bit collision cases, token collision cases:", totals)


if __name__ == "__main__":
    main()
