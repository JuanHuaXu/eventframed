"""Freeze a public near-duplicate event corpus and separate query splits."""

import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path
from urllib.request import urlopen


ROOT = Path(__file__).parent / "usgs-headroom-v1"
SOURCE = ROOT / "source.geojson"
URL = (
    "https://earthquake.usgs.gov/fdsnws/event/1/query?format=geojson"
    "&starttime=2019-07-04&endtime=2019-07-08"
    "&minlatitude=35.4&maxlatitude=36.2"
    "&minlongitude=-118.2&maxlongitude=-117.2"
    "&minmagnitude=4&orderby=time-asc"
)


def write_new(name, value):
    with (ROOT / name).open("x") as handle:
        json.dump(value, handle, indent=2)
        handle.write("\n")


def stamp(feature):
    return datetime.fromtimestamp(feature["properties"]["time"] / 1000, timezone.utc)


def main():
    ROOT.mkdir(exist_ok=True)
    if not SOURCE.exists():
        with urlopen(URL, timeout=30) as response:
            source = response.read()
        with SOURCE.open("xb") as handle:
            handle.write(source)
    source_bytes = SOURCE.read_bytes()
    source = json.loads(source_bytes)
    assert source["type"] == "FeatureCollection" and len(source["features"]) >= 30
    features = sorted(source["features"], key=lambda e: e["properties"]["time"])
    assert len({f["id"] for f in features}) == len(features)
    by_day = {
        day: [f for f in features if stamp(f).strftime("%Y-%m-%d") == day]
        for day in ("2019-07-04", "2019-07-05", "2019-07-06", "2019-07-07")
    }
    assert len(by_day["2019-07-04"]) >= 12 and len(by_day["2019-07-06"]) >= 12
    assert by_day["2019-07-05"] and by_day["2019-07-07"]
    selected = [("design", f) for f in by_day["2019-07-04"][:12]]
    selected += [("confirmation", f) for f in by_day["2019-07-06"][:12]]
    assert len(selected) == 24
    corpus, queries, oracle = [], [], {}
    for split, feature in selected:
        p = feature["properties"]
        at = stamp(feature)
        iso = at.strftime("%Y-%m-%dT%H:%M:%SZ")
        natural = at.strftime("%B %d, %Y at %H:%M:%S UTC")
        event_id = feature["id"]
        assert p["mag"] is not None and p["url"].startswith("https://earthquake.usgs.gov/")
        corpus.append({
            "fixture_id": event_id,
            "text": f"USGS earthquake at {iso}; location {p['place']}; magnitude {p['mag']} {p['magType']}.",
            "source": p["url"],
        })
        for wording, question in (
            ("literal", f"What magnitude did USGS report for the earthquake at {iso}?"),
            ("paraphrase", f"How strong was the earthquake recorded on {natural}?"),
        ):
            case = event_id + "-" + wording
            assert str(p["mag"]) not in question
            queries.append({"case_id": case, "question": question, "split": split})
            oracle[case] = {"target": event_id, "answer": p["mag"], "split": split, "wording": wording}
    for feature in (by_day["2019-07-05"][0], by_day["2019-07-07"][0]):
        iso = stamp(feature).strftime("%Y-%m-%dT%H:%M:%SZ")
        case = "absent-" + feature["id"]
        queries.append({"case_id": case, "question": f"What magnitude did USGS report for the earthquake at {iso}?", "split": "design"})
        oracle[case] = {"target": None, "answer": None, "split": "design", "wording": "absent"}
    assert len(corpus) == 24 and len(queries) == 50 and len(oracle) == 50
    write_new("corpus.json", corpus)
    write_new("queries.json", queries)
    write_new("oracle.json", oracle)
    write_new("source-metadata.json", {
        "query_url": URL,
        "source_sha256": hashlib.sha256(source_bytes).hexdigest(),
        "event_count": len(features),
        "design_ids": [f["id"] for split, f in selected if split == "design"],
        "confirmation_ids": [f["id"] for split, f in selected if split == "confirmation"],
        "absent_ids": [by_day["2019-07-05"][0]["id"], by_day["2019-07-07"][0]["id"]],
    })


if __name__ == "__main__":
    main()
