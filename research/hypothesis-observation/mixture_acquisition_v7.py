"""One-step acquisition under the full joint reliability mixture."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import random
import statistics
import reliability_v6 as previous

base = previous.base
CASES = previous.CASES
ARMS = previous.ARMS + ['acquired']


def gain(model, t):
    mass = [[0.0]*4 for _ in range(2)]
    for weight, component in zip(model.weights, model.models):
        for h, (p, q) in enumerate(zip(component.w, component.row(t))):
            mass[0][h % 4] += weight*p*(1-q)
            mass[1][h % 4] += weight*p*q
    expected = sum(sum(p*p for p in row)/sum(row) for row in mass if sum(row) > 0)
    return expected-sum(p*p for p in model.forecast())


def select(model, seen):
    pairs = [(t, s) for t in range(8) for s in range(4) if (t, s) not in seen]
    if not pairs:
        return None
    scores = {t: gain(model, t) for t in {t for t, _ in pairs}}
    return max(pairs, key=lambda p: (round(scores[p[0]], 14), -p[0], -p[1]))


def episode(case, seed, calibration):
    result = previous.episode(case, seed, calibration)
    table = base.old.likelihood('noise05' if case.endswith('05') else 'noise20')
    model = previous.Reliability(table, result['signals'], previous.likelihoods(calibration))
    rng = random.Random(seed*10+1)
    tape = [[rng.random() for _ in range(8)] for _ in range(4)]
    trace, seen = [], set()
    for _ in range(16):
        t, s = select(model, seen)
        # Only the environment reads latent truth, after the selector has returned.
        y = tape[s if result['modes'][t] else 0][t] < table[t][result['truth']]
        trace.append(dict(forecast=model.forecast(), pair=(t, s), outcome=y,
                          expected_gain=gain(model, t), mode_weights=model.weights[:]))
        model.observe(t, y)
        seen.add((t, s))
    final = model.forecast()
    truth = result['truth'] % 4
    chosen = max(range(4), key=final.__getitem__)
    result['arms']['acquired'] = dict(trace=trace, final=final,
        curve_brier=statistics.mean(base.old.brier(x['forecast'], truth) for x in trace),
        final_brier=base.old.brier(final, truth), correct=chosen == truth,
        confident_wrong=max(final) >= .9 and chosen != truth, cost=8+len(seen),
        mode_weights=model.weights[:])
    return result


def summarize(records):
    summaries, contrasts, failures = [], [], []
    for split in range(2):
        for case in CASES:
            rows = [r for r in records if r['split'] == split and r['case'] == case]
            for arm in ARMS:
                summaries.append(dict(split=split, case=case, arm=arm, n=len(rows),
                    **{m: statistics.mean(r['arms'][arm][m] for r in rows) for m in
                       ['curve_brier', 'final_brier', 'correct', 'confident_wrong', 'cost']}))
            for control in previous.ARMS:
                for metric in ['curve_brier', 'final_brier']:
                    ds = [r['arms'][control][metric]-r['arms']['acquired'][metric] for r in rows]
                    m = statistics.mean(ds)
                    rad = 3.3*statistics.stdev(ds)/math.sqrt(len(ds))
                    contrasts.append(dict(split=split, case=case, control=control, metric=metric,
                                          gain=m, lower=m-rad, upper=m+rad))
                    if split != 1:
                        continue
                    if (control == 'fixed' or (control == 'reliable' and case == 'independent20')) and m < -.01:
                        failures.append(f'{case} {control} {metric} harm')
                    if control == 'reliable' and case == 'matched_misleading20' and metric == 'final_brier':
                        if m < .03 or m-rad <= 0:
                            failures.append('misleading-signal rescue')
                    if control == 'average' and case == 'independent20' and metric == 'curve_brier':
                        if m < .005 or m-rad <= 0:
                            failures.append('acquisition speed rescue')
    return dict(summaries=summaries, contrasts=contrasts, failures=failures, screen_passed=not failures)


def main():
    root = Path(__file__).parent
    calibration = previous.previous.prior.calibrate()
    records = []
    for split in range(2):
        for j, case in enumerate(CASES):
            for i in range(128):
                r = episode(case, 202609170417+split*100000+j*1000+i, calibration)
                r['split'] = split
                records.append(r)
    result = dict(records=records, calibration=calibration, **summarize(records),
        hashes={p: hashlib.sha256((root/p).read_bytes()).hexdigest() for p in
                ['mixture_acquisition_v7.py', 'MIXTURE_ACQUISITION_PROTOCOL.md', 'reliability_v6.py',
                 'factorial_v5.py', 'informed_v4.py', 'dependence_v3.py', 'experiment.py']})
    with (root/'mixture-acquisition-v7.json.gz').open('xb') as f, gzip.GzipFile(fileobj=f, mode='wb', mtime=0) as z:
        z.write(json.dumps(result).encode())
    result.pop('records')
    with (root/'mixture-acquisition-v7-summary.json').open('x') as f:
        json.dump(result, f, indent=2)
    print(json.dumps(dict(screen_passed=result['screen_passed'], failures=result['failures'])))


if __name__ == '__main__':
    main()
