"""Costed fresh-measurement research; freshness is a declared model assumption."""
import json
import math
import random
import statistics
import experiment
from dependence_v3 import Joint

CASES = ['independent20', 'copied20', 'mixed20', 'copied05', 'false_renewal20']
ARMS = ['regular', 'mixed', 'random', 'entropy']


class Observer:
    def __init__(self, table):
        self.model = Joint(table)
        self.used = [[0] * 8, [0] * 8]

    def actions(self, budget, regular_only=False):
        return [(kind, t, self.used[kind][t])
                for kind in range(1 if regular_only else 2)
                for t in range(8)
                if kind+1 <= budget and self.used[kind][t] < (4 if kind == 0 else 8)]

    def row(self, action):
        kind, t, _ = action
        return self.model.row(t) if kind == 0 else self.model.table[t][:]

    def forecast(self):
        return experiment.classes(self.model.w, 'noise05')

    def observe(self, action, outcome):
        kind, t, slot = action
        assert slot == self.used[kind][t]
        if kind == 0:
            self.model.observe(t, outcome)
        else:
            # A new measurement is not another report from the ordinary root.
            # It changes h's posterior, leaving P(mode_t | h, old reports) intact.
            self.model.w = experiment.update(self.model.w, self.row(action), outcome)
        self.used[kind][t] += 1
        assert abs(sum(self.model.w) - 1) < 1e-12

    def risk(self):
        return 1 - sum(p*p for p in self.forecast())

    def clone(self):
        # Likelihood tables are immutable; only posterior and paid-slot state
        # need branch-local ownership during counterfactual planning.
        child = object.__new__(Observer)
        child.model = object.__new__(Joint)
        child.model.table = self.model.table
        child.model.w = self.model.w[:]
        child.model.mode = [row[:] for row in self.model.mode]
        child.model.first = self.model.first[:]
        child.used = [row[:] for row in self.used]
        return child

    def branch_risk(self, action, credits, regular_only):
        row = self.row(action)
        p = sum(w*q for w, q in zip(self.model.w, row))
        result = 0.0
        for outcome, mass in ((False, 1-p), (True, p)):
            child = self.clone()
            child.observe(action, outcome)
            remaining = credits - (action[0]+1)
            if remaining == 0:
                value = child.risk()
            else:
                value = min(child.branch_risk(a, remaining, regular_only)
                            for a in child.actions(remaining, regular_only))
            result += mass * value
        return result

    def select(self, budget, arm, rng):
        actions = self.actions(budget, arm == 'regular')
        if arm == 'random':
            return rng.choice(actions)
        best, chosen = -math.inf, None
        for action in actions:
            if arm == 'entropy':
                p = sum(w*q for w, q in zip(self.model.w, self.row(action)))
                score = experiment.entropy([p, 1-p]) / (action[0]+1)
            else:
                if action[0]+1 > min(2, budget):
                    continue
                score = -self.branch_risk(action, min(2, budget), arm == 'regular')
            if score > best + 1e-14:
                best, chosen = score, action
        assert chosen is not None
        return chosen


def episode(case, seed):
    truth = random.Random(seed*10).randrange(16)
    rng = random.Random(seed*10+1)
    ordinary = [[rng.random() for _ in range(8)] for _ in range(4)]
    rng = random.Random(seed*10+3)
    fresh = [[rng.random() for _ in range(8)] for _ in range(8)]
    table = experiment.likelihood('noise05' if case.endswith('05') else 'noise20')
    arms = {}
    for arm in ARMS:
        observer, rng = Observer(table), random.Random(seed*10+2)
        trace, spent, area = [], 0, 0.0
        while spent < 16:
            forecast = observer.forecast()
            action = observer.select(16-spent, arm, rng)
            kind, t, slot = action
            copied = case in ('copied20', 'copied05', 'false_renewal20') or (case == 'mixed20' and t % 2 == 0)
            # Simulator truth/tapes are accessed only after policy selection.
            if kind == 0:
                draw = ordinary[0 if copied else slot][t]
            else:
                draw = ordinary[0][t] if case == 'false_renewal20' else fresh[slot][t]
            outcome = draw < table[t][truth]
            cost = kind+1
            area += cost * experiment.brier(forecast, truth % 4) / 16
            trace.append(dict(credit=spent, action=action, cost=cost,
                              forecast=forecast, outcome=outcome))
            observer.observe(action, outcome)
            spent += cost
        final = observer.forecast()
        chosen = max(range(4), key=final.__getitem__)
        assert spent == 16 and len({tuple(t['action']) for t in trace}) == len(trace)
        arms[arm] = dict(trace=trace, final=final,
                         final_brier=experiment.brier(final, truth % 4),
                         credit_brier=area, correct=chosen == truth % 4,
                         confident_wrong=max(final) >= .9 and chosen != truth % 4,
                         renewals=sum(t['action'][0] == 1 for t in trace),
                         calls=len(trace), credits=spent)
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
                       for m in ('final_brier', 'credit_brier', 'correct', 'confident_wrong', 'renewals', 'calls')}))
            for control in ('regular', 'random', 'entropy'):
                ds = [r['arms'][control]['final_brier']-r['arms']['mixed']['final_brier'] for r in rs]
                mean = statistics.mean(ds)
                radius = 3.3 * statistics.stdev(ds) / math.sqrt(64)
                required = (case in ('copied20', 'mixed20') or
                            (control == 'regular' and case in ('independent20', 'copied05')))
                passed = (mean >= (.02 if control == 'regular' else 0) and mean-radius > 0) if case in ('copied20', 'mixed20') else mean >= -.01
                gates.append(dict(split=split, case=case, control=control,
                                  gain=mean, lower=mean-radius, upper=mean+radius,
                                  required=required, passed=passed))
    return dict(summaries=summaries, comparisons=gates,
                screen_passed=all(g['passed'] for g in gates if g['required']))


def main():
    records = []
    for split in range(2):
        for j, case in enumerate(CASES):
            for i in range(64):
                seed = 2027110701*1000000 + split*100000 + j*1000 + i
                r = episode(case, seed)
                r['split'] = split
                records.append(r)
    print(json.dumps(dict(records=records, **summarize(records))))


if __name__ == '__main__':
    main()
