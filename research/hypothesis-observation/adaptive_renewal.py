"""Closed-loop uncertain measurement origins; isolated finite research only."""
import json
import math
import random
import statistics
import experiment
from renewal import Observer, CASES
from uncertain_renewal import UncertainRenewal

ARMS = ['regular', 'certain_mixed', 'uncertain_regular', 'uncertain_mixed',
        'uncertain_random', 'uncertain_entropy']


class AdaptiveObserver(Observer):
    def __init__(self, table):
        self.model = UncertainRenewal(table)
        self.used = self.model.used

    def row(self, action):
        return self.model.row(action)

    def observe(self, action, outcome):
        # Branches share immutable latent rows until an observation replaces
        # the selected type's rows. Never mutate a sibling branch's posterior.
        t = action[1]
        self.model.local[t] = self.model.local[t][:]
        self.model.observe(action, outcome)

    def clone(self):
        child = object.__new__(AdaptiveObserver)
        child.model = object.__new__(UncertainRenewal)
        child.model.table = self.model.table
        child.model.w = self.model.w[:]
        child.model.local = self.model.local[:]
        child.model.used = [row[:] for row in self.used]
        child.used = child.model.used
        return child


def policy(arm):
    return arm.replace('uncertain_', '').replace('certain_', '')


def episode(case, seed):
    truth = random.Random(seed*10).randrange(16)
    rng = random.Random(seed*10+1)
    ordinary = [[rng.random() for _ in range(8)] for _ in range(4)]
    rng = random.Random(seed*10+3)
    fresh = [[rng.random() for _ in range(8)] for _ in range(8)]
    table = experiment.likelihood('noise05' if case.endswith('05') else 'noise20')
    arms = {}
    for arm in ARMS:
        observer = AdaptiveObserver(table) if arm.startswith('uncertain_') else Observer(table)
        rng, trace, spent, area = random.Random(seed*10+2), [], 0, 0.0
        while spent < 16:
            forecast = observer.forecast()
            action = observer.select(16-spent, policy(arm), rng)
            kind, t, slot = action
            copied = case in ('copied20', 'copied05', 'false_renewal20') or (case == 'mixed20' and t % 2 == 0)
            if kind == 0:
                draw = ordinary[0 if copied else slot][t]
            else:
                draw = ordinary[0][t] if case == 'false_renewal20' else fresh[slot][t]
            outcome, cost = draw < table[t][truth], kind+1
            area += cost * experiment.brier(forecast, truth % 4) / 16
            trace.append(dict(credit=spent, action=action, cost=cost, forecast=forecast, outcome=outcome))
            observer.observe(action, outcome)
            spent += cost
        final = observer.forecast()
        chosen = max(range(4), key=final.__getitem__)
        assert spent == 16 and len({tuple(s['action']) for s in trace}) == len(trace)
        arms[arm] = dict(trace=trace, final=final,
            final_brier=experiment.brier(final, truth % 4), credit_brier=area,
            correct=chosen == truth % 4,
            confident_wrong=max(final) >= .9 and chosen != truth % 4,
            renewals=sum(s['action'][0] == 1 for s in trace), calls=len(trace), credits=spent)
        if isinstance(observer, AdaptiveObserver):
            arms[arm]['freshness'] = [observer.model.fresh_probability(t) for t in range(8)]
    return dict(case=case, seed=seed, truth=truth, arms=arms)


def summarize(records):
    summaries, gates = [], []
    for split in range(2):
        for case in CASES:
            rs = [r for r in records if r['split'] == split and r['case'] == case]
            assert len(rs) == 64
            for arm in ARMS:
                summaries.append(dict(split=split, case=case, arm=arm,
                    **{m: statistics.mean(r['arms'][arm][m] for r in rs)
                       for m in ['final_brier', 'credit_brier', 'correct', 'confident_wrong', 'renewals', 'calls']}))
            tests = []
            if case != 'false_renewal20':
                tests += [('nonharm', c) for c in ['regular', 'certain_mixed']]
            else:
                tests += [('gain', 'certain_mixed'), ('nonharm', 'regular')]
            if case in ('copied20', 'mixed20'):
                tests += [('gain', c) for c in ['regular', 'uncertain_random', 'uncertain_entropy']]
            for kind, control in tests:
                ds = [r['arms'][control]['final_brier']-r['arms']['uncertain_mixed']['final_brier'] for r in rs]
                mean, radius = statistics.mean(ds), 3.3 * statistics.stdev(ds) / 8
                minimum = .02 if control in ('regular', 'certain_mixed') else 0
                passed = mean >= minimum and mean-radius > 0 if kind == 'gain' else mean-radius >= -.01
                gates.append(dict(split=split, case=case, control=control, kind=kind,
                                  gain=mean, lower=mean-radius, upper=mean+radius, passed=passed))
    assert len(gates) == 32
    return dict(summaries=summaries, gates=gates, screen_passed=all(g['passed'] for g in gates))


def main():
    records = []
    for split in range(2):
        for j, case in enumerate(CASES):
            for i in range(64):
                r = episode(case, 2027111801*1000000 + split*100000 + j*1000 + i)
                r['split'] = split
                records.append(r)
    print(json.dumps(dict(records=records, **summarize(records))))


if __name__ == '__main__':
    main()
