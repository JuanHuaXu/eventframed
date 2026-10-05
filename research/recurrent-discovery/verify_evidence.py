"""Reject incomplete/mismatched pilot artifacts; print all paired comparisons."""

import argparse
import hashlib
import json
import math
from pathlib import Path

from experiment import CONFIG, SCHEDULES, make_data, signature


def verify(path):
    data = json.loads(path.read_text())
    root = Path(__file__).parent
    assert data["complete"] is True
    assert data["config"] == CONFIG
    for key, filename in [("source_sha256", "experiment.py"),
                          ("protocol_sha256", "PROTOCOL.md")]:
        assert data[key] == hashlib.sha256((root / filename).read_bytes()).hexdigest()
    expected = {(seed, task, arm) for seed in CONFIG["seeds"]
                for task in ["addition", "random_labels"] for arm in SCHEDULES}
    runs = {(r["seed"], r["task"], r["arm"]): r for r in data["runs"]}
    assert set(runs) == expected and len(data["runs"]) == len(expected)
    for (seed, task, arm), r in runs.items():
        _, _, partition = make_data(seed, task, CONFIG["modulus"], CONFIG["train_fraction"])
        assert r["partition"] == partition
        assert r["checkpoint_reload_equal"] is True
        assert r["initial_parameter_hash"] != r["final_parameter_hash"]
        assert r["schedule"] == SCHEDULES[arm]
        steps = CONFIG["updates"] * (3 if arm == "single_extended" else 1)
        assert r["steps"] == steps
        assert r["core_calls"] == steps * len(SCHEDULES[arm])
        assert [c["step"] for c in r["curve"]] == list(range(0, steps + 1, CONFIG["checkpoint_every"]))
        assert r["final"] == r["curve"][-1]
        assert r["signature"] == signature(r["curve"])
        for c in r["curve"]:
            assert c["core_calls"] == c["step"] * len(SCHEDULES[arm])
            for split in ["train", "test"]:
                assert 0 <= c[split]["accuracy"] <= 1
                assert math.isfinite(c[split]["log_loss"]) and c[split]["log_loss"] >= 0
    paired = []
    for seed in CONFIG["seeds"]:
        for task in ["addition", "random_labels"]:
            group = [runs[seed, task, arm] for arm in SCHEDULES]
            assert len({r["initial_parameter_hash"] for r in group}) == 1
            assert len({r["parameters"] for r in group}) == 1
            nested = runs[seed, task, "nested"]["final"]["test"]
            for control in ["flat", "single_extended"]:
                other = runs[seed, task, control]["final"]["test"]
                paired.append(dict(seed=seed, task=task, control=control,
                                   accuracy_gain=nested["accuracy"] - other["accuracy"],
                                   log_loss_reduction=other["log_loss"] - nested["log_loss"]))
    structured = [p for p in paired if p["task"] == "addition"]
    promising = (all(p["accuracy_gain"] >= 0.05 and p["log_loss_reduction"] > 0
                     for p in structured) and
                 all(r["final"]["test"]["accuracy"] <= 0.10 for r in data["runs"]
                     if r["task"] == "random_labels"))
    return dict(verified_runs=len(runs), promising_signal=promising, paired=paired,
                delayed_signatures=sum(r["signature"]["delayed_generalization_signature"]
                                       for r in data["runs"]))


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("artifact", type=Path)
    args = parser.parse_args()
    print(json.dumps(verify(args.artifact), indent=2))
