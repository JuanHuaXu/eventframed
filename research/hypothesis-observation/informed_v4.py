"""Disjoint synthetic provenance calibration for the finite source-mode model."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import random
import statistics
import dependence_v3 as base

CASES = base.CASES + ['signal_random20', 'signal_misleading20']
ARMS = ['joint_gini', 'informed', 'independent', 'once']


def calibrate():
    rng = random.Random(731946281)
    counts = [[0, 0], [0, 0]]
    for _ in range(4096):
        mode = int(rng.random() < .5)
        signal = mode if rng.random() < .9 else 1-mode
        counts[mode][signal] += 1
    likelihood = [[(n+.5)/(sum(row)+1) for n in row] for row in counts]
    priors = [likelihood[1][s]/(likelihood[0][s]+likelihood[1][s]) for s in range(2)]
    return dict(counts=counts, priors=priors)


def episode(case, seed, calibration):
    truth = random.Random(seed*10).randrange(16)
    rng = random.Random(seed*10+1)
    tape = [[rng.random() for _ in range(8)] for _ in range(4)]
    table = base.old.likelihood('noise05' if case.endswith('05') else 'noise20')
    modes = [not (case.startswith('copied') or
             ((case == 'mixed20' or case.startswith('signal_')) and t % 2 == 0)) for t in range(8)]
    reliability = .5 if case == 'signal_random20' else .1 if case == 'signal_misleading20' else .9
    srng = random.Random(seed*10+3)
    signals = [int(m if srng.random() < reliability else not m) for m in modes]
    result = dict(case=case, seed=seed, truth=truth, signals=signals, arms={})
    for arm in ARMS:
        model = base.Joint(table, arm in ('independent', 'once'))
        if arm == 'informed':
            model.mode = [[calibration['priors'][s]]*16 for s in signals]
        seen = set()
        trace = []
        rng = random.Random(seed*10+2)
        for _ in range(16):
            forecast = base.old.classes(model.w, 'noise05')
            pair = model.select(seen, arm, rng)
            outcome = None
            if pair is not None:
                t, s = pair
                outcome = tape[s if modes[t] else 0][t] < table[t][truth]
                model.observe(t, outcome)
                seen.add(pair)
            trace.append(dict(forecast=forecast, pair=pair, outcome=outcome))
        final = base.old.classes(model.w, 'noise05')
        chosen = max(range(4), key=final.__getitem__)
        result['arms'][arm] = dict(trace=trace, final=final,
            curve_brier=statistics.mean(base.old.brier(x['forecast'], truth % 4) for x in trace),
            final_brier=base.old.brier(final, truth % 4), correct=chosen == truth % 4,
            confident_wrong=max(final) >= .9 and chosen != truth % 4,
            queries=len(seen), cost=8+len(seen))
    return result


def summarize(records):
    summaries, comparisons, failures = [], [], []
    for split in range(2):
        for case in CASES:
            rows = [r for r in records if r['split'] == split and r['case'] == case]
            for arm in ARMS:
                summaries.append(dict(split=split, case=case, arm=arm, n=len(rows),
                    **{m: statistics.mean(r['arms'][arm][m] for r in rows) for m in
                       ['curve_brier', 'final_brier', 'correct', 'confident_wrong', 'cost']}))
            for control in ['joint_gini', 'independent', 'once']:
                for metric in ['curve_brier', 'final_brier', 'confident_wrong']:
                    ds = [r['arms'][control][metric]-r['arms']['informed'][metric] for r in rows]
                    gain = statistics.mean(ds)
                    radius = 3.3*statistics.stdev(ds)/math.sqrt(len(ds))
                    comparisons.append(dict(split=split, case=case, control=control,
                        metric=metric, gain=gain, lower=gain-radius, upper=gain+radius))
                    if split != 1 or control != 'joint_gini' or case not in base.CASES:
                        continue
                    if metric != 'confident_wrong' and gain < -.01:
                        failures.append(f'{case} {metric} harm')
                    if case == 'independent20' and metric == 'curve_brier' and (gain < .01 or gain-radius <= 0):
                        failures.append('independent20 recovery')
                    if case in ['copied20', 'mixed20'] and metric == 'confident_wrong' and gain < -.02:
                        failures.append(f'{case} confidence harm')
    return dict(summaries=summaries, comparisons=comparisons, failures=failures, screen_passed=not failures)


def main():
    root = Path(__file__).parent
    calibration = calibrate()
    records = []
    for split in range(2):
        for j, case in enumerate(CASES):
            for i in range(128):
                r = episode(case, 202609120417+split*100000+j*1000+i, calibration)
                r['split'] = split
                records.append(r)
    result = dict(records=records, calibration=calibration, **summarize(records),
        hashes={p: hashlib.sha256((root/p).read_bytes()).hexdigest() for p in
                ['informed_v4.py', 'INFORMED_PROTOCOL.md', 'dependence_v3.py', 'experiment.py']})
    with (root/'informed-v4.json.gz').open('xb') as f, gzip.GzipFile(fileobj=f, mode='wb', mtime=0) as z:
        z.write(json.dumps(result).encode())
    result.pop('records')
    with (root/'informed-v4-summary.json').open('x') as f:
        json.dump(result, f, indent=2)
    print(json.dumps(dict(screen_passed=result['screen_passed'], failures=result['failures'])))


if __name__ == '__main__':
    main()
