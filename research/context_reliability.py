"""Bounded, past-delivered context reliability on consumed v19 traces."""
from collections import OrderedDict, deque
import hashlib
import json
from pathlib import Path
import sys

import prequential_reliability as previous


class ContextReliability:
    def __init__(self):
        self.history = OrderedDict()

    def status(self, key):
        history = self.history.get(key, ())
        if len(history) < 8:
            return "unknown"
        nominal = sum(p for _, p in history) / len(history)
        correct = sum(c for c, _ in history)
        rate = (correct + 8 * nominal) / (len(history) + 8)
        return "warn" if rate < nominal - .03 else "clear"

    def observe(self, key, correct, nominal):
        if key not in self.history:
            if len(self.history) == 512:
                self.history.popitem(last=False)
            self.history[key] = deque(maxlen=64)
        self.history[key].append((correct, nominal))
        self.history.move_to_end(key)


def key_at(record, step):
    last = record["Views"][step][2]["trace"][-1]
    mask, values = last["observed"], last["values"]
    assert 0 <= mask < 512 and 0 <= values < 512 and values & ~mask == 0
    p = record["Ticks"][step]["Predictions"][2]["P"]
    return mask, values, int(p >= .5)


def replay_record(record):
    state = ContextReliability()
    records, seen = [], set()
    for step, tick in enumerate(record["Ticks"]):
        if previous.confidence_record(record, step):
            key = key_at(record, step)
            # Outcomes here are metric targets only; the decision uses history.
            records.append(dict(step=step, status=state.status(key),
                                error=key[2] != tick["Outcome"]))
        for origin in tick.get("Delivered") or []:
            assert 0 <= origin <= step and origin not in seen
            assert not record["Ticks"][origin]["Missing"]
            seen.add(origin)
            if previous.confidence_record(record, origin):
                key = key_at(record, origin)
                earlier = record["Ticks"][origin]
                p = earlier["Predictions"][2]["P"]
                state.observe(key, key[2] == earlier["Outcome"], max(p, 1-p))
    return records


def summarize(path):
    # Reuse the frozen aggregation/gate, restoring the hook even on failure.
    original = previous.replay_record
    try:
        previous.replay_record = replay_record
        result = previous.summarize(path)
    finally:
        previous.replay_record = original
    for group in result["groups"]:
        counts = group["counts"]
        total = sum(v["frames"] for v in counts.values())
        group["unknown_fraction"] = counts.get("unknown", {}).get("frames", 0) / total if total else None
    return result


if __name__ == "__main__":
    result = summarize(Path(sys.argv[1]))
    sources = {p: Path(p).read_text() for p in
               ("research/context_reliability.py", "research/prequential_reliability.py")}
    result["sources"] = sources
    result["hashes"] = {p: hashlib.sha256(s.encode()).hexdigest() for p, s in sources.items()}
    with Path(sys.argv[2]).open("x") as f:
        json.dump(result, f, indent=2)
        f.write("\n")
    print(json.dumps(dict(passed=result["passed"], target=result["target"]), indent=2))
