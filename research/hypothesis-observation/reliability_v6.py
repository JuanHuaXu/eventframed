"""Exact finite averaging over global provenance reliability modes."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import statistics
import factorial_v5 as previous

base = previous.base
CASES = previous.CASES
ARMS = ['fixed', 'reliable', 'average']


def likelihoods(calibration):
    measured = [[(n+.5)/(sum(row)+1) for n in row] for row in calibration['counts']]
    return [measured, [[.5, .5], [.5, .5]], [list(reversed(row)) for row in measured]]


class Reliability:
    def __init__(self, table, signals, matrices):
        self.models = []
        self.weights = []
        for matrix in matrices:
            model = base.Joint(table)
            weight = 1/len(matrices)
            for t, signal in enumerate(signals):
                mass = (matrix[0][signal]+matrix[1][signal])/2
                model.mode[t] = [matrix[1][signal]/(2*mass)]*16
                weight *= mass
            self.models.append(model)
            self.weights.append(weight)
        self.weights = [w/sum(self.weights) for w in self.weights]

    def hypothesis(self):
        return [sum(w*m.w[h] for w, m in zip(self.weights, self.models)) for h in range(16)]

    def forecast(self):
        return base.old.classes(self.hypothesis(), 'noise05')

    def observe(self, t, y):
        weights = []
        for w, model in zip(self.weights, self.models):
            p = sum(h*q for h, q in zip(model.w, model.row(t)))
            weights.append(w*(p if y else 1-p))
            model.observe(t, y)
        self.weights = [w/sum(weights) for w in weights]


def episode(case, seed, calibration):
    original = previous.episode(case, seed, calibration)
    table = base.old.likelihood('noise05' if case.endswith('05') else 'noise20')
    model = Reliability(table, original['signals'], likelihoods(calibration))
    trace = []
    for step in original['arms']['FF']['trace']:
        trace.append(dict(forecast=model.forecast(), pair=step['pair'], outcome=step['outcome'],
                          mode_weights=model.weights[:]))
        model.observe(step['pair'][0], step['outcome'])
    final = model.forecast()
    truth = original['truth']
    chosen = max(range(4), key=final.__getitem__)
    average = dict(trace=trace, final=final,
        curve_brier=statistics.mean(base.old.brier(x['forecast'], truth % 4) for x in trace),
        final_brier=base.old.brier(final, truth % 4), correct=chosen == truth % 4,
        confident_wrong=max(final) >= .9 and chosen != truth % 4, cost=24,
        mode_weights=model.weights[:])
    return dict(case=case, seed=seed, truth=truth, signals=original['signals'], modes=original['modes'],
                arms=dict(fixed=original['arms']['FF'], reliable=original['arms']['FI'], average=average))


def summarize(records):
    summaries, contrasts, failures = [], [], []
    for split in range(2):
        for case in CASES:
            rows = [r for r in records if r['split'] == split and r['case'] == case]
            for arm in ARMS:
                summaries.append(dict(split=split, case=case, arm=arm, n=len(rows),
                    **{m: statistics.mean(r['arms'][arm][m] for r in rows) for m in
                       ['curve_brier', 'final_brier', 'correct', 'confident_wrong', 'cost']},
                    mode_weights=[statistics.mean(r['arms'][arm]['mode_weights'][i] for r in rows)
                                  for i in range(3)] if arm == 'average' else None))
            for control in ['fixed', 'reliable']:
                for metric in ['curve_brier', 'final_brier']:
                    ds = [r['arms'][control][metric]-r['arms']['average'][metric] for r in rows]
                    gain = statistics.mean(ds)
                    rad = 3.3*statistics.stdev(ds)/math.sqrt(len(ds))
                    contrasts.append(dict(split=split, case=case, control=control, metric=metric,
                                          gain=gain, lower=gain-rad, upper=gain+rad))
                    if split != 1:
                        continue
                    if (control == 'fixed' or case == 'independent20') and gain < -.01:
                        failures.append(f'{case} {control} {metric} harm')
                    if control == 'reliable' and case == 'matched_misleading20' and metric == 'final_brier':
                        if gain < .03 or gain-rad <= 0:
                            failures.append('misleading-signal rescue')
    return dict(summaries=summaries, contrasts=contrasts, failures=failures, screen_passed=not failures)


def main():
    root = Path(__file__).parent
    calibration = previous.prior.calibrate()
    records = []
    for split in range(2):
        for j, case in enumerate(CASES):
            for i in range(128):
                r = episode(case, 202609150173+split*100000+j*1000+i, calibration)
                r['split'] = split
                records.append(r)
    result = dict(records=records, calibration=calibration, **summarize(records),
        hashes={p: hashlib.sha256((root/p).read_bytes()).hexdigest() for p in
                ['reliability_v6.py', 'RELIABILITY_PROTOCOL.md', 'factorial_v5.py',
                 'informed_v4.py', 'dependence_v3.py', 'experiment.py']})
    with (root/'reliability-v6.json.gz').open('xb') as f, gzip.GzipFile(fileobj=f, mode='wb', mtime=0) as z:
        z.write(json.dumps(result).encode())
    result.pop('records')
    with (root/'reliability-v6-summary.json').open('x') as f:
        json.dump(result, f, indent=2)
    print(json.dumps(dict(screen_passed=result['screen_passed'], failures=result['failures'])))


if __name__ == '__main__':
    main()
