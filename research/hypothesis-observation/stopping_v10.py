"""Fixed-acquisition prefix stopping; forecast accuracy and cost reported separately."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import random
import statistics
import time
import lookahead_rollout_v9 as previous

base = previous.previous
CASES = previous.CASES
ARMS = ['lookahead', 'stop_one', 'stop_two']
PRICE = .005


def stop_now(one, two, remaining, cautious):
    return one <= PRICE and (not cautious or remaining == 1 or two <= 2*PRICE)


def arm(trace, final, truth, stop=16):
    forecast = trace[stop]['forecast'][:] if stop < 16 else final[:]
    emitted = [dict(forecast=x['forecast'][:], pair=x['pair'], outcome=x['outcome'])
               if i < stop else dict(forecast=forecast[:], pair=None, outcome=None)
               for i, x in enumerate(trace)]
    chosen = max(range(4), key=forecast.__getitem__)
    loss = base.base.old.brier(forecast, truth)
    return dict(trace=emitted, final=forecast, stop_at=stop, reports=stop, cost=8+stop,
                final_brier=loss, penalized=loss+PRICE*(8+stop),
                curve_brier=statistics.mean(base.base.old.brier(x['forecast'], truth) for x in emitted),
                correct=chosen == truth, confident_wrong=max(forecast) >= .9 and chosen != truth)


def episode(case, seed, calibration):
    result = base.episode(case, seed, calibration)
    table = base.base.old.likelihood('noise05' if case.endswith('05') else 'noise20')
    model = base.previous.Reliability(table, result['signals'], base.previous.likelihoods(calibration))
    rng = random.Random(seed*10+1)
    tape = [[rng.random() for _ in range(8)] for _ in range(4)]
    seen, trace, decisions = set(), [], []
    stops = dict(stop_one=16, stop_two=16)
    for step in range(16):
        remaining = 16-step
        if remaining > 1:
            inspected = previous.planner.inspect(model, seen)
            one = max(x['one'] for x in inspected['scores'])
            two = max(x['two'] for x in inspected['scores'])
            pair = inspected['lookahead']
        else:
            pair = base.select(model, seen)
            one, two = base.gain(model, pair[0]), None
        decisions.append(dict(one=one, two=two))
        for name, cautious in [('stop_one', False), ('stop_two', True)]:
            if stops[name] == 16 and stop_now(one, two, remaining, cautious):
                stops[name] = step
        t, s = pair
        forecast = model.forecast()
        y = tape[s if result['modes'][t] else 0][t] < table[t][result['truth']]
        trace.append(dict(pair=pair, forecast=forecast, outcome=y))
        model.observe(t, y)
        seen.add(pair)
    final, truth = model.forecast(), result['truth'] % 4
    result['arms']['lookahead'] = arm(trace, final, truth)
    for name in stops:
        result['arms'][name] = arm(trace, final, truth, stops[name])
    result['decisions'] = decisions
    return result


def summarize(records):
    summaries, contrasts, failures = [], [], []
    for split in range(2):
        for case in CASES:
            rows = [r for r in records if r['split'] == split and r['case'] == case]
            for name in ARMS:
                summaries.append(dict(split=split, case=case, arm=name, n=len(rows),
                    **{m: statistics.mean(r['arms'][name][m] for r in rows) for m in
                       ['curve_brier', 'final_brier', 'penalized', 'reports', 'cost', 'correct', 'confident_wrong']}))
            for name in ['stop_one', 'stop_two']:
                for metric in ['curve_brier', 'final_brier', 'penalized', 'reports']:
                    ds = [r['arms']['lookahead'][metric]-r['arms'][name][metric] for r in rows]
                    gain = statistics.mean(ds)
                    rad = 3.3*statistics.stdev(ds)/math.sqrt(len(ds))
                    contrasts.append(dict(split=split, case=case, arm=name, metric=metric,
                                          gain=gain, lower=gain-rad, upper=gain+rad))
                    if split != 1 or name != 'stop_two':
                        continue
                    if metric in ['curve_brier', 'final_brier'] and gain < -.01:
                        failures.append(f'{case} {metric} harm')
                    if case in ['copied20', 'matched05']:
                        if metric == 'reports' and gain < 3.2:
                            failures.append(f'{case} report savings')
                        if metric == 'penalized' and gain-rad <= 0:
                            failures.append(f'{case} penalized benefit')
    return dict(summaries=summaries, contrasts=contrasts, failures=failures, screen_passed=not failures)


def main():
    root = Path(__file__).parent
    calibration = base.previous.previous.prior.calibrate()
    records = []
    start = time.perf_counter()
    for split in range(2):
        for j, case in enumerate(CASES):
            for i in range(128):
                r = episode(case, 202609220137+split*100000+j*1000+i, calibration)
                r['split'] = split
                records.append(r)
            print('completed split', split, case, flush=True)
    result = dict(records=records, calibration=calibration, **summarize(records),
        elapsed_seconds=time.perf_counter()-start,
        hashes={p: hashlib.sha256((root/p).read_bytes()).hexdigest() for p in
                ['stopping_v10.py', 'STOPPING_PROTOCOL.md', 'lookahead_rollout_v9.py', 'lookahead_v8.py',
                 'mixture_acquisition_v7.py', 'reliability_v6.py', 'factorial_v5.py',
                 'informed_v4.py', 'dependence_v3.py', 'experiment.py']})
    with (root/'stopping-v10.json.gz').open('xb') as f, gzip.GzipFile(fileobj=f, mode='wb', mtime=0) as z:
        z.write(json.dumps(result).encode())
    result.pop('records')
    with (root/'stopping-v10-summary.json').open('x') as f:
        json.dump(result, f, indent=2)
    print(json.dumps(dict(screen_passed=result['screen_passed'], failures=result['failures'])))


if __name__ == '__main__':
    main()
