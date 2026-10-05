"""Finite population check of shared/local freshness on fixed paid reports."""
import itertools
import json
import experiment
from renewal import Observer
from uncertain_renewal import UncertainRenewal
from hierarchical_renewal import HierarchicalRenewal

SCHEDULE = [(0, 0, 0), (0, 1, 0), (0, 2, 0), (0, 7, 0),
            (1, 0, 0), (1, 1, 0), (1, 0, 1), (1, 1, 1), (1, 2, 0), (1, 7, 0)]
CASES = ['genuine', 'counterfeit', 'even_counterfeit', 'odd_counterfeit']


def main():
    table = experiment.likelihood('noise20')
    names = ['certain', 'local', 'hierarchical']
    sums = {case: dict(mass=0., oracle_brier=0.,
            models={name: dict(brier=0., correct=0., confident_wrong=0.) for name in names})
            for case in CASES}
    for outcomes in itertools.product((False, True), repeat=10):
        models = [Observer(table), UncertainRenewal(table), HierarchicalRenewal(table)]
        for action, outcome in zip(SCHEDULE, outcomes):
            for model in models:
                model.observe(action, outcome)
        forecasts = [m.forecast() for m in models]
        for case in CASES:
            joint = []
            for h in range(16):
                w, roots = 1/16, {}
                for (kind, t, _), y in zip(SCHEDULE, outcomes):
                    q = table[t][h]
                    if kind == 0:
                        roots[t] = y
                        w *= q if y else 1-q
                    else:
                        copied = (case == 'counterfeit' or
                                  case == 'even_counterfeit' and t % 2 == 0 or
                                  case == 'odd_counterfeit' and t % 2 == 1)
                        w *= int(y == roots[t]) if copied else q if y else 1-q
                joint.append(w)
            mass = sum(joint)
            sums[case]['mass'] += mass
            if mass == 0:
                continue
            true_forecast = experiment.classes([w/mass for w in joint], 'noise20')
            sums[case]['oracle_brier'] += sum(w*experiment.brier(true_forecast, h % 4) for h, w in enumerate(joint))
            for name, p in zip(names, forecasts):
                chosen = max(range(4), key=p.__getitem__)
                out = sums[case]['models'][name]
                out['brier'] += sum(w*experiment.brier(p, h % 4) for h, w in enumerate(joint))
                out['correct'] += sum(w for h, w in enumerate(joint) if h % 4 == chosen)
                if max(p) >= .9:
                    out['confident_wrong'] += sum(w for h, w in enumerate(joint) if h % 4 != chosen)
    for r in sums.values():
        assert abs(r['mass']-1) < 1e-12
        for m in r['models'].values():
            assert m['brier'] >= r['oracle_brier']-1e-12
    gates = []
    for case in CASES:
        r = sums[case]['models']
        gain = r['local']['brier']-r['hierarchical']['brier']
        gates.append(dict(case=case, kind='nonharm', gain=gain, passed=gain >= -.01))
    r = sums['genuine']['models']
    gain = r['local']['brier']-r['hierarchical']['brier']
    gates.append(dict(case='genuine', kind='gain', gain=gain, passed=gain >= .005))
    r = sums['counterfeit']['models']
    gain = r['certain']['confident_wrong']-r['hierarchical']['confident_wrong']
    gates.append(dict(case='counterfeit', kind='false_confidence_reduction', gain=gain, passed=gain >= .05))
    print(json.dumps(dict(schedule=SCHEDULE, credits=sum(a[0]+1 for a in SCHEDULE),
                         results=sums, gates=gates, screen_passed=all(g['passed'] for g in gates)), indent=2))


if __name__ == '__main__':
    main()
