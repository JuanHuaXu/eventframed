"""Materialize a frozen public retrieval transfer set from declarative facts."""

import json
from pathlib import Path


ROOT = Path(__file__).parent
SET = ROOT / "nara-fusion-v1"
NASA = ROOT / "rfc-transfer-v1" / "corpus.json"


def write(name, value):
    (SET / name).write_text(json.dumps(value, indent=2) + "\n")


def main():
    facts = json.loads((SET / "facts.json").read_text())
    nasa = json.loads(NASA.read_text())[:13]
    assert len(facts) == 15 and len(nasa) == 13
    assert len({f["fixture_id"] for f in facts}) == 15
    corpus = list(nasa)
    queries = []
    oracle = {}
    for f in facts:
        corpus.append({
            "fixture_id": f["fixture_id"],
            "text": f"National Archives milestone: {f['title']} ({f['year']}); {f['description']}.",
            "source": f["source"],
        })
        for wording, question in (
            ("literal", f"What year does the National Archives list for {f['title']}?"),
            ("paraphrase", f["paraphrase"]),
        ):
            case = f["fixture_id"] + "-" + wording
            assert f["year"] not in question
            queries.append({"case_id": case, "question": question})
            oracle[case] = {"target": f["fixture_id"], "answer": f["year"], "wording": wording}
    for case, question in (
        ("nara-absent-emancipation", "What year does the Archives list for the Emancipation Proclamation?"),
        ("nara-absent-suffrage", "What year does the Archives list for the women's voting rights amendment?"),
    ):
        queries.append({"case_id": case, "question": question})
        oracle[case] = {"target": None, "answer": None, "wording": "absent"}
    assert len(corpus) == 28 and len(queries) == 32
    write("corpus.json", corpus)
    write("queries.json", queries)
    write("oracle.json", oracle)


if __name__ == "__main__":
    main()
