"""Fresh fixed-report-budget rollout of exact receding two-step acquisition."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import random
import statistics
import time
import lookahead_v8 as planner

previous = planner.previous
CASES = previous.CASES
ARMS = previous.ARMS + ['lookahead']


def action(model, seen, remaining):
    if remaining == 1:
        return previous.select(model, seen)
    return planner.inspect(model, seen)['lookahead']


def episode(case, seed, calibration):
    result = previous.episode(case, seed, calibration)
    table = previous.base.old.likelihood('noise05' if case.endswith('05') else 'noise20')
    model = previous.previous.Reliability(table, result['signals'], previous.previous.likelihoods(calibration))
    rng = random.Random(seed*10+1)
    tape = [[rng.random() for _ in range(8)] for _ in range(4)]
    seen, trace = set(), []
    for step in range(16):
        t, s = action(model, seen, 16-step)
        forecast = model.forecast()
        y = tape[s if result['modes'][t] else 0][t] < table[t][result['truth']]
        trace.append(dict(pair=(t, s), forecast=forecast, outcome=y, mode_weights=model.weights[:]))
        model.observe(t, y)
        seen.add((t, s))
    final = model.forecast()
    truth = result['truth'] % 4
    chosen = max(range(4), key=final.__getitem__)
    result['arms']['lookahead'] = dict(trace=trace, final=final,
        curve_brier=statistics.mean(previous.base.old.brier(x['forecast'], truth) for x in trace),
        final_brier=previous.base.old.brier(final, truth), correct=chosen == truth,
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
                    ds = [r['arms'][control][metric]-r['arms']['lookahead'][metric] for r in rows]
                    gain = statistics.mean(ds)
                    rad = 3.3*statistics.stdev(ds)/math.sqrt(len(ds))
                    contrasts.append(dict(split=split, case=case, control=control, metric=metric,
                                          gain=gain, lower=gain-rad, upper=gain+rad))
                    if split != 1:
                        continue
                    if (control == 'fixed' or (control == 'reliable' and case == 'independent20')) and gain < -.01:
                        failures.append(f'{case} {control} {metric} harm')
                    if control == 'reliable' and case == 'matched_misleading20' and metric == 'final_brier':
                        if gain < .03 or gain-rad <= 0:
                            failures.append('misleading-signal rescue')
                    if control == 'acquired' and case == 'independent20' and metric == 'curve_brier':
                        if gain < .005 or gain-rad <= 0:
                            failures.append('lookahead speed rescue')
    return dict(summaries=summaries, contrasts=contrasts, failures=failures, screen_passed=not failures)


def main():
    root = Path(__file__).parent
    calibration = previous.previous.previous.prior.calibrate()
    records = []
    start = time.perf_counter()
    for split in range(2):
        for j, case in enumerate(CASES):
            for i in range(128):
                r = episode(case, 202609190613+split*100000+j*1000+i, calibration)
                r['split'] = split
                records.append(r)
            print('completed split', split, case, flush=True)
    result = dict(records=records, calibration=calibration, **summarize(records),
        elapsed_seconds=time.perf_counter()-start,
        hashes={p: hashlib.sha256((root/p).read_bytes()).hexdigest() for p in
                ['lookahead_rollout_v9.py', 'LOOKAHEAD_ROLLOUT_PROTOCOL.md', 'lookahead_v8.py',
                 'mixture_acquisition_v7.py', 'reliability_v6.py', 'factorial_v5.py',
                 'informed_v4.py', 'dependence_v3.py', 'experiment.py']})
    with (root/'lookahead-rollout-v9.json.gz').open('xb') as f, gzip.GzipFile(fileobj=f, mode='wb', mtime=0) as z:
        z.write(json.dumps(result).encode())
    result.pop('records')
    with (root/'lookahead-rollout-v9-summary.json').open('x') as f:
        json.dump(result, f, indent=2)
    print(json.dumps(dict(screen_passed=result['screen_passed'], failures=result['failures'])))


if __name__ == '__main__':
    main()
