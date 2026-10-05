"""Exact stress over fixed-budget allocations and all counterfeit masks."""
import itertools
import json
import experiment
from renewal import Observer
from uncertain_renewal import UncertainRenewal
from hierarchical_renewal import HierarchicalRenewal

TYPES = [0, 1, 2, 7]


def evaluate(counts):
    table = experiment.likelihood('noise20')
    schedule = [(0, t, 0) for t in TYPES] + [(1, t, slot) for t, n in zip(TYPES, counts) for slot in range(n)]
    assert len(schedule) == 10 and sum(a[0]+1 for a in schedule) == 16
    assert len(set(schedule)) == 10
    sums = [dict(mask=mask, mass=0., oracle=0., brier=[0., 0., 0.], wrong=[0., 0., 0.]) for mask in range(16)]
    for outcomes in itertools.product((False, True), repeat=10):
        models = [Observer(table), UncertainRenewal(table), HierarchicalRenewal(table)]
        for action, y in zip(schedule, outcomes):
            for model in models:
                model.observe(action, y)
        forecasts = [m.forecast() for m in models]
        norms = [sum(p*p for p in ps) for ps in forecasts]
        chosen = [max(range(4), key=ps.__getitem__) for ps in forecasts]
        for mask, out in enumerate(sums):
            weights = []
            for h in range(16):
                w = 1/16
                for i, (t, n) in enumerate(zip(TYPES, counts)):
                    root, q = outcomes[i], table[t][h]
                    w *= q if root else 1-q
                    start = 4 + sum(counts[:i])
                    for y in outcomes[start:start+n]:
                        w *= int(y == root) if mask & (1 << i) else q if y else 1-q
                weights.append(w)
            mass = sum(weights)
            out['mass'] += mass
            if mass == 0:
                continue
            cls = [sum(w for h, w in enumerate(weights) if h % 4 == c) for c in range(4)]
            out['oracle'] += mass - sum(w*w/mass for w in cls)
            for a, ps in enumerate(forecasts):
                out['brier'][a] += sum(w*(1+norms[a]-2*ps[h % 4]) for h, w in enumerate(weights))
                if max(ps) >= .9:
                    out['wrong'][a] += mass-cls[chosen[a]]
    for s in sums:
        assert abs(s['mass']-1) < 1e-12
        assert all(x >= s['oracle']-1e-12 for x in s['brier'])
    return dict(counts=counts, cells=sums)


def main():
    allocations = [list(c) for c in itertools.product(range(1, 4), repeat=4) if sum(c) == 6]
    assert len(allocations) == 10
    results = [evaluate(c) for c in allocations]
    gates = []
    for r in results:
        for s in r['cells']:
            gain = s['brier'][1]-s['brier'][2]
            gates.append(dict(counts=r['counts'], mask=s['mask'], kind='nonharm', gain=gain, passed=gain >= -.01))
            if s['mask'] == 0:
                gates.append(dict(counts=r['counts'], mask=0, kind='gain', gain=gain, passed=gain >= .005))
            if s['mask'] == 15:
                gain = s['wrong'][0]-s['wrong'][2]
                gates.append(dict(counts=r['counts'], mask=15, kind='false_confidence', gain=gain, passed=gain >= .05))
    assert len(gates) == 180
    print(json.dumps(dict(types=TYPES, results=results, gates=gates,
                         screen_passed=all(g['passed'] for g in gates)), indent=2))


if __name__ == '__main__':
    main()
