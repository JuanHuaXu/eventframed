"""Freeze a Puerto Rico earthquake corpus with UTC/local-time query pairs."""

import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path
from urllib.request import urlopen
from zoneinfo import ZoneInfo


ROOT = Path(__file__).parent / "pr-time-v1"
SOURCE = ROOT / "source.geojson"
URL = (
    "https://earthquake.usgs.gov/fdsnws/event/1/query?format=geojson"
    "&starttime=2020-01-07&endtime=2020-01-09"
    "&minlatitude=17.4&maxlatitude=18.5"
    "&minlongitude=-67.8&maxlongitude=-65.9"
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
    features = sorted(source["features"], key=lambda e: e["properties"]["time"])
    assert source["type"] == "FeatureCollection" and len(features) >= 26
    assert len({f["id"] for f in features}) == len(features)
    selected = features[:24]
    assert len({int(f["properties"]["time"] / 1000) for f in selected}) == 24
    local_zone = ZoneInfo("America/Puerto_Rico")
    corpus, queries, oracle = [], [], {}
    for i, feature in enumerate(selected):
        p = feature["properties"]
        at = stamp(feature)
        iso = at.strftime("%Y-%m-%dT%H:%M:%SZ")
        local = at.astimezone(local_zone)
        natural = local.strftime("%B %d, %Y at %I:%M:%S %p")
        assert local.utcoffset().total_seconds() == -4 * 3600
        event_id = feature["id"]
        split = "design" if i < 12 else "confirmation"
        assert p["mag"] is not None and p["url"].startswith("https://earthquake.usgs.gov/")
        corpus.append({
            "fixture_id": event_id,
            "text": f"USGS Puerto Rico earthquake at {iso}; location {p['place']}; magnitude {p['mag']} {p['magType']}.",
            "source": p["url"],
        })
        for wording, question in (
            ("utc", f"What magnitude did USGS report for the earthquake at {iso}?"),
            ("local", f"How strong was the Puerto Rico earthquake on {natural} Atlantic Standard Time?"),
        ):
            case = event_id + "-" + wording
            assert f"magnitude {p['mag']}" not in question
            assert iso not in question or wording == "utc"
            queries.append({"case_id": case, "question": question, "split": split})
            oracle[case] = {"target": event_id, "answer": p["mag"], "split": split, "wording": wording}
    for feature in features[24:26]:
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
        "design_ids": [f["id"] for f in selected[:12]],
        "confirmation_ids": [f["id"] for f in selected[12:]],
        "absent_ids": [f["id"] for f in features[24:26]],
        "local_zone": "America/Puerto_Rico",
    })


if __name__ == "__main__":
    main()
