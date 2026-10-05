"""Freeze a broad official USGS event set and labels before retrieval."""

import hashlib
import json
import math
from datetime import datetime, timedelta, timezone
from pathlib import Path
from urllib.request import urlopen


ROOT = Path(__file__).parent / "usgs-broad-confirmation-v1"
SOURCE = ROOT / "source.geojson"
URL = (
    "https://earthquake.usgs.gov/fdsnws/event/1/query?format=geojson"
    "&starttime=2026-09-15&endtime=2026-09-23"
    "&minmagnitude=4&orderby=time-asc&limit=2000"
)


def write_new(path, value):
    with path.open("x") as handle:
        json.dump(value, handle, indent=2)
        handle.write("\n")


def main():
    ROOT.mkdir(exist_ok=True)
    if SOURCE.exists():
        source_bytes = SOURCE.read_bytes()
    else:
        with urlopen(URL, timeout=30) as response:
            source_bytes = response.read()
        captured_at = datetime.now(timezone.utc).isoformat()
        with SOURCE.open("xb") as handle:
            handle.write(source_bytes)
        with (ROOT / "captured-at.txt").open("x") as handle:
            handle.write(captured_at + "\n")
    captured_at = (ROOT / "captured-at.txt").read_text().strip()
    source = json.loads(source_bytes)
    assert source["type"] == "FeatureCollection"
    assert len(source["features"]) <= 2000
    eligible = sorted(
        (feature for feature in source["features"]
         if isinstance(feature["properties"]["mag"], (int, float))
         and math.isfinite(feature["properties"]["mag"])
         and isinstance(feature["properties"]["place"], str)
         and feature["properties"]["place"]
         and feature["properties"]["url"].startswith("https://earthquake.usgs.gov/")),
        key=lambda feature: (feature["properties"]["time"], feature["id"]),
    )
    assert len(eligible) >= 24
    selected = eligible[:24]
    assert len({feature["id"] for feature in selected}) == 24
    corpus, queries, oracle = [], [], {}
    for index, feature in enumerate(selected):
        p = feature["properties"]
        assert p["mag"] is not None and p["url"].startswith("https://earthquake.usgs.gov/")
        at = datetime.fromtimestamp(p["time"] / 1000, timezone.utc)
        iso = at.strftime("%Y-%m-%dT%H:%M:%SZ")
        natural = at.strftime("%B %d, %Y at %H:%M:%S UTC")
        offset = (at + timedelta(hours=11)).strftime("%B %d, %Y at %H:%M:%S")
        event_id = feature["id"]
        corpus.append({
            "fixture_id": event_id,
            "text": f"USGS earthquake at {iso}; location {p['place']}; magnitude {p['mag']} {p['magType']}.",
            "source": p["url"],
        })
        split = "design" if index < 12 else "confirmation"
        wordings = (
            ("literal", f"What magnitude did USGS report for the earthquake at {iso}?"),
            ("paraphrase", f"How strong was the earthquake reported at {p['place']} on {natural}?"),
            ("offset", f"What magnitude did USGS list for the earthquake reported at {p['place']} at {offset} in UTC+11 (eleven hours ahead of UTC)?"),
        )
        for wording, question in wordings:
            case = event_id + "-" + wording
            assert f"magnitude {p['mag']}" not in question.lower()
            queries.append({"case_id": case, "question": question, "split": split})
            oracle[case] = {"target": event_id, "answer": p["mag"], "split": split, "wording": wording}
    assert len(corpus) == 24 and len(queries) == len(oracle) == 72
    write_new(ROOT / "corpus.json", corpus)
    write_new(ROOT / "queries.json", queries)
    write_new(ROOT / "oracle.json", oracle)
    write_new(ROOT / "source-metadata.json", {
        "query_url": URL,
        "source_sha256": hashlib.sha256(source_bytes).hexdigest(),
        "captured_at": captured_at,
        "source_feature_count": len(source["features"]),
        "eligible_feature_count": len(eligible),
        "selected_ids": [feature["id"] for feature in selected],
    })


if __name__ == "__main__":
    main()
