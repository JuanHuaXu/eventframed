"""Freeze a fresh official USGS sequence and answer labels before retrieval."""

import hashlib
import json
from datetime import datetime, timedelta, timezone
from pathlib import Path
from urllib.request import urlopen


ROOT = Path(__file__).parent / "usgs-tadine-transfer-v1"
SOURCE = ROOT / "source.geojson"
URL = (
    "https://earthquake.usgs.gov/fdsnws/event/1/query?format=geojson"
    "&starttime=2026-09-23&endtime=2026-10-01"
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
        with SOURCE.open("xb") as handle:
            handle.write(source_bytes)
    source = json.loads(source_bytes)
    assert source["type"] == "FeatureCollection"
    assert len(source["features"]) <= 2000
    selected = sorted(
        (feature for feature in source["features"]
         if "Tadine, New Caledonia" in feature["properties"]["place"]),
        key=lambda feature: (feature["properties"]["time"], feature["id"]),
    )
    assert len(selected) >= 24
    selected = selected[:24]
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
            ("paraphrase", f"How strong was the Tadine-area earthquake recorded on {natural}?"),
            ("offset", f"What magnitude did USGS list for the Tadine-area earthquake at {offset} in UTC+11 (eleven hours ahead of UTC)?"),
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
        "source_feature_count": len(source["features"]),
        "matching_feature_count": len([feature for feature in source["features"] if "Tadine, New Caledonia" in feature["properties"]["place"]]),
        "selected_ids": [feature["id"] for feature in selected],
    })


if __name__ == "__main__":
    main()
