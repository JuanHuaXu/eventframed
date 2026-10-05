"""Isolated recurrent-learning pilot. Nothing here is imported by eventframed."""

import argparse
import hashlib
import importlib.metadata
import json
import math
from pathlib import Path
import platform
import random
import time

import torch
from torch import nn
from torch.nn import functional as F


CONFIG = dict(modulus=31, train_fraction=0.6, seeds=[2026091201, 2026091202],
              updates=3000, checkpoint_every=100, embedding=32, hidden=64,
              learning_rate=0.001, weight_decay=1.0, betas=[0.9, 0.98], threads=1)
SCHEDULES = dict(single="LH", flat="LHLHLH", nested="LLHLLH",
                 single_extended="LH")


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True).encode()).hexdigest()


def make_data(seed, task, p=31, fraction=0.6):
    """Only this generator sees arithmetic; the learner receives opaque indices."""
    if task not in {"addition", "random_labels"}:
        raise ValueError("unknown task")
    rng = random.Random(seed)
    permutation = list(range(p))
    rng.shuffle(permutation)
    groups = [(a, b) for a in range(p) for b in range(a, p)]
    rng.shuffle(groups)
    cut = int(len(groups) * fraction)
    # Use a separate stream so label randomization cannot change the partition.
    labels = random.Random(seed + 100000000)
    train, test = [], []
    for i, (a, b) in enumerate(groups):
        y = permutation[(a + b) % p] if task == "addition" else labels.randrange(p)
        destination = train if i < cut else test
        for x, z in [(a, b)] if a == b else [(a, b), (b, a)]:
            destination.append([permutation[x], permutation[z], y])
    return torch.tensor(train), torch.tensor(test), {
        "train_hash": digest(train), "test_hash": digest(test),
        "train_groups": cut, "test_groups": len(groups) - cut,
        "train_examples": len(train), "test_examples": len(test)}


class Learner(nn.Module):
    def __init__(self, p=31, embedding=32, hidden=64):
        super().__init__()
        self.embedding = nn.Embedding(p, embedding)
        self.input = nn.Linear(2 * embedding, hidden)
        self.low = nn.Linear(3 * hidden, hidden)
        self.high = nn.Linear(3 * hidden, hidden)
        self.output = nn.Linear(hidden, p)

    def forward(self, pairs, schedule):
        x = torch.tanh(self.input(self.embedding(pairs).flatten(1)))
        low, high = torch.zeros_like(x), torch.zeros_like(x)
        # Every forward resets state. Repetition changes computation, not evidence.
        for module in schedule:
            state = torch.cat((low, high, x), dim=1)
            if module == "L":
                low = torch.tanh(self.low(state))
            elif module == "H":
                high = torch.tanh(self.high(state))
            else:
                raise ValueError("unknown recurrent module")
        return self.output(high)


def update(model, optimizer, train, schedule):
    # There is intentionally no held-out argument or generator inside this API.
    optimizer.zero_grad(set_to_none=True)
    loss = F.cross_entropy(model(train[:, :2], schedule), train[:, 2])
    if not torch.isfinite(loss):
        raise RuntimeError("nonfinite training loss")
    loss.backward()
    optimizer.step()


@torch.no_grad()
def evaluate(model, data, schedule):
    logits = model(data[:, :2], schedule)
    return {"accuracy": (logits.argmax(1) == data[:, 2]).double().mean().item(),
            "log_loss": F.cross_entropy(logits, data[:, 2]).item()}


def signature(curve):
    def sustained(field, threshold, earliest=0):
        for i in range(len(curve) - 2):
            if curve[i]["step"] >= earliest and all(
                    c[field]["accuracy"] >= threshold for c in curve[i:i + 3]):
                return curve[i]["step"]
        return None
    fit = sustained("train", 0.99)
    general = sustained("test", 0.90)
    # Generalization must FIRST appear after fitting, not merely remain high later.
    delayed = fit is not None and general is not None and general >= fit + 500
    return {"fit_step": fit, "generalization_step": general,
            "delayed_generalization_signature": delayed}


def parameter_hash(model):
    h = hashlib.sha256()
    for name, value in model.state_dict().items():
        h.update(name.encode())
        h.update(value.detach().cpu().numpy().tobytes())
    return h.hexdigest()


@torch.no_grad()
def inference_timing(model, pair, schedule):
    for _ in range(20):
        model(pair, schedule)
    samples = []
    for _ in range(200):
        start = time.perf_counter_ns()
        model(pair, schedule)
        samples.append((time.perf_counter_ns() - start) / 1000)
    samples.sort()
    return {"unit": "microseconds", "n": len(samples),
            "median": samples[len(samples) // 2],
            "p95": samples[math.ceil(0.95 * len(samples)) - 1], "max": samples[-1]}


def run(seed, task, arm, checkpoint_dir):
    train, test, partition = make_data(seed, task, CONFIG["modulus"], CONFIG["train_fraction"])
    torch.manual_seed(seed)
    model = Learner(CONFIG["modulus"], CONFIG["embedding"], CONFIG["hidden"])
    initial_hash = parameter_hash(model)
    optimizer = torch.optim.AdamW(model.parameters(), lr=CONFIG["learning_rate"],
                                 weight_decay=CONFIG["weight_decay"],
                                 betas=tuple(CONFIG["betas"]))
    schedule = SCHEDULES[arm]
    steps = CONFIG["updates"] * (3 if arm == "single_extended" else 1)
    curve, training_seconds = [], 0.0
    for step in range(steps + 1):
        if step % CONFIG["checkpoint_every"] == 0:
            curve.append({"step": step, "core_calls": step * len(schedule),
                          "train": evaluate(model, train, schedule),
                          "test": evaluate(model, test, schedule),
                          "parameter_l2": math.sqrt(sum(p.detach().square().sum().item()
                                                       for p in model.parameters()))})
        if step < steps:
            start = time.perf_counter()
            update(model, optimizer, train, schedule)
            training_seconds += time.perf_counter() - start
    # Persistent parameter learning is checked independently of a serving publish.
    path = checkpoint_dir / f"{seed}-{task}-{arm}.pt"
    torch.save(model.state_dict(), path)
    restored = Learner(CONFIG["modulus"], CONFIG["embedding"], CONFIG["hidden"])
    restored.load_state_dict(torch.load(path, weights_only=True))
    with torch.no_grad():
        reload_equal = torch.equal(model(test[:, :2], schedule),
                                   restored(test[:, :2], schedule))
    if not reload_equal:
        raise RuntimeError("checkpoint round-trip changed predictions")
    final_hash = parameter_hash(model)
    doubled = evaluate(model, test, schedule * 2)
    if final_hash != parameter_hash(model):
        raise RuntimeError("inference mutated persistent parameters")
    return dict(seed=seed, task=task, arm=arm, schedule=schedule, partition=partition,
                parameters=sum(p.numel() for p in model.parameters()),
                initial_parameter_hash=initial_hash, final_parameter_hash=final_hash,
                steps=steps, core_calls=steps * len(schedule), curve=curve,
                signature=signature(curve), final=curve[-1],
                frozen_double_loop_test=doubled, checkpoint_reload_equal=reload_equal,
                training_seconds=training_seconds,
                inference=inference_timing(model, test[:1, :2], schedule))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--checkpoints", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        parser.error("refusing to overwrite existing experiment evidence")
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.checkpoints.mkdir(parents=True, exist_ok=True)
    torch.set_num_threads(CONFIG["threads"])
    torch.use_deterministic_algorithms(True)
    root = Path(__file__).parent
    result = dict(protocol="research/recurrent-discovery/PROTOCOL.md", config=CONFIG,
                  source_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
                  protocol_sha256=hashlib.sha256((root / "PROTOCOL.md").read_bytes()).hexdigest(),
                  python=platform.python_version(), torch=torch.__version__,
                  packages={name: importlib.metadata.version(name) for name in
                            ["numpy", "torch", "sympy", "networkx", "filelock", "fsspec",
                             "jinja2", "MarkupSafe", "mpmath", "typing-extensions"]},
                  platform=platform.platform(), device="cpu", complete=False, runs=[])
    # Incremental aggregate artifacts preserve failed/interrupted runs as incomplete.
    for seed in CONFIG["seeds"]:
        for task in ["addition", "random_labels"]:
            for arm in SCHEDULES:
                row = run(seed, task, arm, args.checkpoints)
                result["runs"].append(row)
                args.output.write_text(json.dumps(result, indent=2, allow_nan=False) + "\n")
                print(json.dumps({k: row[k] for k in ["seed", "task", "arm", "final",
                                                     "signature", "training_seconds"]}), flush=True)
    result["complete"] = True
    args.output.write_text(json.dumps(result, indent=2, allow_nan=False) + "\n")


if __name__ == "__main__":
    main()
