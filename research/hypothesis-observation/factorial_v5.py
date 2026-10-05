"""Separate what provenance makes us observe from how it weights observations."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import random
import statistics
import informed_v4 as prior

base = prior.base
CASES = ['independent20', 'copied20', 'mixed20', 'matched05', 'matched20',
         'matched_random20', 'matched_misleading20']
ARMS = ['FF', 'FI', 'IF', 'II']
CONTRASTS = [('FF', 'IF'), ('FF', 'FI'), ('FI', 'II'), ('IF', 'II'), ('FF', 'II')]


def episode(case, seed, calibration):
    truth = random.Random(seed*10).randrange(16)
    rng = random.Random(seed*10+1)
    tape = [[rng.random() for _ in range(8)] for _ in range(4)]
    table = base.old.likelihood('noise05' if case.endswith('05') else 'noise20')
    mrng = random.Random(seed*10+4)
    modes = [mrng.random() < .5 for _ in range(8)] if case.startswith('matched') else [
        not (case.startswith('copied') or (case == 'mixed20' and t % 2 == 0)) for t in range(8)]
    accuracy = .5 if 'random' in case else .1 if 'misleading' in case else .9
    srng = random.Random(seed*10+3)
    signals = [int(m if srng.random() < accuracy else not m) for m in modes]
    result = dict(case=case, seed=seed, truth=truth, modes=modes, signals=signals, arms={})
    for acquisition in ['F', 'I']:
        scorers = {s: base.Joint(table) for s in ['F', 'I']}
        scorers['I'].mode = [[calibration['priors'][s]]*16 for s in signals]
        traces = {s: [] for s in scorers}
        seen = set()
        rng = random.Random(seed*10+2)
        for _ in range(16):
            forecasts = {s: base.old.classes(model.w, 'noise05') for s, model in scorers.items()}
            t, slot = scorers[acquisition].select(seen, 'joint_gini', rng)
            y = tape[slot if modes[t] else 0][t] < table[t][truth]
            for s, model in scorers.items():
                traces[s].append(dict(forecast=forecasts[s], pair=(t, slot), outcome=y))
                model.observe(t, y)
            seen.add((t, slot))
        for s, model in scorers.items():
            final = base.old.classes(model.w, 'noise05')
            chosen = max(range(4), key=final.__getitem__)
            result['arms'][acquisition+s] = dict(trace=traces[s], final=final,
                curve_brier=statistics.mean(base.old.brier(x['forecast'], truth % 4) for x in traces[s]),
                final_brier=base.old.brier(final, truth % 4), correct=chosen == truth % 4,
                confident_wrong=max(final) >= .9 and chosen != truth % 4,
                cost=8+len(seen), distinct_tests=len({t for t, _ in seen}))
    return result


def summarize(records):
    summaries, contrasts = [], []
    for split in range(2):
        for case in CASES:
            rows = [r for r in records if r['split'] == split and r['case'] == case]
            for arm in ARMS:
                summaries.append(dict(split=split, case=case, arm=arm, n=len(rows),
                    **{m: statistics.mean(r['arms'][arm][m] for r in rows) for m in
                       ['curve_brier', 'final_brier', 'correct', 'confident_wrong', 'cost', 'distinct_tests']}))
            for control, treatment in CONTRASTS:
                for metric in ['curve_brier', 'final_brier']:
                    ds = [r['arms'][control][metric]-r['arms'][treatment][metric] for r in rows]
                    gain = statistics.mean(ds)
                    rad = 3.3*statistics.stdev(ds)/math.sqrt(len(ds))
                    contrasts.append(dict(split=split, case=case, control=control, treatment=treatment,
                        metric=metric, gain=gain, lower=gain-rad, upper=gain+rad,
                        direction='improvement' if gain-rad > 0 else 'harm' if gain+rad < 0 else 'inconclusive'))
    return dict(summaries=summaries, contrasts=contrasts)


def main():
    root = Path(__file__).parent
    calibration = prior.calibrate()
    records = []
    for split in range(2):
        for j, case in enumerate(CASES):
            for i in range(128):
                r = episode(case, 202609131901+split*100000+j*1000+i, calibration)
                r['split'] = split
                records.append(r)
    result = dict(records=records, calibration=calibration, **summarize(records),
        hashes={p: hashlib.sha256((root/p).read_bytes()).hexdigest() for p in
                ['factorial_v5.py', 'FACTORIAL_PROTOCOL.md', 'informed_v4.py', 'dependence_v3.py', 'experiment.py']})
    with (root/'factorial-v5.json.gz').open('xb') as f, gzip.GzipFile(fileobj=f, mode='wb', mtime=0) as z:
        z.write(json.dumps(result).encode())
    result.pop('records')
    with (root/'factorial-v5-summary.json').open('x') as f:
        json.dump(result, f, indent=2)
    print('episodes:', len(records), 'diagnostic only; no adoption gate')


if __name__ == '__main__':
    main()
