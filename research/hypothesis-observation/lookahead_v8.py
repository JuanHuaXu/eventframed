"""Model-implied two-step value, never empirical forecast benefit."""
import copy
import gzip
import hashlib
import json
from pathlib import Path
import statistics
import time
import mixture_acquisition_v7 as previous


def candidates(seen):
    return [(t, next(s for s in range(4) if (t, s) not in seen))
            for t in range(8) if any((t, s) not in seen for s in range(4))]


def probability(model, t):
    return sum(w*sum(h*q for h, q in zip(m.w, m.row(t)))
               for w, m in zip(model.weights, model.models))


def values(model, seen):
    out = []
    for pair in candidates(seen):
        t, _ = pair
        value = previous.gain(model, t)
        p = probability(model, t)
        remaining = seen | {pair}
        for y, chance in [(False, 1-p), (True, p)]:
            if chance <= 0:
                continue
            child = copy.deepcopy(model)
            child.observe(t, y)
            gains = [previous.gain(child, tt) for tt, _ in candidates(remaining)]
            value += chance*max(gains, default=0.0)
        out.append(dict(pair=pair, one=previous.gain(model, t), two=value))
    return out


def inspect(model, seen):
    scores = values(model, seen)
    if not scores:
        return dict(scores=[], greedy=None, lookahead=None, value_gain=0.0, changed=False)
    choose = lambda field: max(scores, key=lambda x: (round(x[field], 14), -x['pair'][0], -x['pair'][1]))
    greedy, look = choose('one'), choose('two')
    value_gain = look['two']-greedy['two']
    return dict(scores=scores, greedy=greedy['pair'], lookahead=look['pair'],
                value_gain=value_gain, changed=greedy['pair'] != look['pair'] and value_gain > 1e-10)


def run_records(parent):
    records = []
    for case in previous.CASES:
        episodes = sorted((r for r in parent['records'] if r['split'] == 0 and r['case'] == case), key=lambda r: r['seed'])[:16]
        for r in episodes:
            table = previous.base.old.likelihood('noise05' if case.endswith('05') else 'noise20')
            model = previous.previous.Reliability(table, r['signals'], previous.previous.likelihoods(parent['calibration']))
            seen = set()
            for step, report in enumerate(r['arms']['acquired']['trace']):
                if step in [0, 4, 8, 12]:
                    if any(abs(a-b) > 1e-12 for a, b in zip(model.forecast(), report['forecast'])):
                        raise ValueError('prefix forecast mismatch')
                    records.append(dict(case=case, seed=r['seed'], prefix=step, **inspect(model, seen)))
                    if step == 12:
                        break
                model.observe(report['pair'][0], report['outcome'])
                seen.add(tuple(report['pair']))
    return records


def summarize(records):
    cases, prefixes = [], []
    def summary(rows):
        return dict(states=len(rows), episodes=len({r['seed'] for r in rows}),
                    mean_gain=statistics.mean(r['value_gain'] for r in rows),
                    max_gain=max(r['value_gain'] for r in rows),
                    changed=sum(r['changed'] for r in rows)/len(rows))
    for case in previous.CASES:
        cases.append(dict(case=case, **summary([r for r in records if r['case'] == case])))
        for prefix in [0, 4, 8, 12]:
            prefixes.append(dict(case=case, prefix=prefix, **summary([r for r in records if r['case'] == case and r['prefix'] == prefix])))
    warranted = any(r['case'] in ['independent20', 'matched_misleading20'] and
                    r['mean_gain'] >= .002 and r['changed'] >= .1 for r in cases)
    return dict(cases=cases, prefixes=prefixes, fresh_rollout_warranted=warranted)


def main():
    root = Path(__file__).parent
    parentpath = root/'mixture-acquisition-v7.json.gz'
    with gzip.open(parentpath, 'rt') as f:
        parent = json.load(f)
    for p, digest in parent['hashes'].items():
        if hashlib.sha256((root/p).read_bytes()).hexdigest() != digest:
            raise ValueError('parent source changed: '+p)
    start = time.perf_counter()
    records = run_records(parent)
    result = dict(records=records, **summarize(records), elapsed_seconds=time.perf_counter()-start,
                  parent_hash=hashlib.sha256(parentpath.read_bytes()).hexdigest(),
                  hashes={p: hashlib.sha256((root/p).read_bytes()).hexdigest() for p in
                          ['lookahead_v8.py', 'LOOKAHEAD_PROTOCOL.md']})
    with (root/'lookahead-v8.json').open('x') as f:
        json.dump(result, f, indent=2)
    print(json.dumps({k: v for k, v in result.items() if k not in ['records', 'prefixes', 'hashes']}))


if __name__ == '__main__':
    main()
