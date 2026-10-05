"""Matched acquired-history comparison, not a new adaptive policy."""
import json
import math
import statistics
import sys
import experiment
from hierarchical_renewal import HierarchicalRenewal
from uncertain_renewal import UncertainRenewal


def run(data):
    records = []
    for r in data['records']:
        table = experiment.likelihood('noise05' if r['case'].endswith('05') else 'noise20')
        arms = {}
        for arm, old in r['arms'].items():
            local, mixture = UncertainRenewal(table), HierarchicalRenewal(table)
            forecasts, area = [], [0., 0.]
            for step in old['trace']:
                ps = [local.forecast(), mixture.forecast()]
                forecasts.append(ps)
                for i, p in enumerate(ps):
                    area[i] += step['cost'] * experiment.brier(p, r['truth'] % 4) / 16
                local.observe(step['action'], step['outcome'])
                mixture.observe(step['action'], step['outcome'])
            ps = [local.forecast(), mixture.forecast()]
            if arm.startswith('uncertain_'):
                assert max(abs(a-b) for a, b in zip(ps[0], old['final'])) < 1e-12
            arms[arm] = dict(forecasts=forecasts, final=ps, credit_brier=area,
                final_brier=[experiment.brier(p, r['truth'] % 4) for p in ps],
                correct=[max(range(4), key=p.__getitem__) == r['truth'] % 4 for p in ps],
                confidently_wrong=[max(p) >= .9 and max(range(4), key=p.__getitem__) != r['truth'] % 4 for p in ps],
                components=mixture.weights(), log_evidence=mixture.log_evidence)
        records.append(dict(seed=r['seed'], split=r['split'], case=r['case'], arms=arms))
    summaries, gates = [], []
    for split in range(2):
        for case in ['independent20', 'copied20', 'mixed20', 'copied05', 'false_renewal20']:
            rs = [r for r in records if r['split'] == split and r['case'] == case]
            assert len(rs) == 64
            for arm in rs[0]['arms']:
                ds = [r['arms'][arm]['final_brier'][0]-r['arms'][arm]['final_brier'][1] for r in rs]
                mean, radius = statistics.mean(ds), 3.3*statistics.stdev(ds)/8
                summaries.append(dict(split=split, case=case, arm=arm,
                    brier=[statistics.mean(r['arms'][arm]['final_brier'][i] for r in rs) for i in (0, 1)],
                    wrong=[statistics.mean(r['arms'][arm]['confidently_wrong'][i] for r in rs) for i in (0, 1)],
                    components=[statistics.mean(r['arms'][arm]['components'][i] for r in rs) for i in (0, 1, 2)],
                    gain=mean, lower=mean-radius, upper=mean+radius))
                if arm == 'uncertain_mixed':
                    gates.append(dict(split=split, case=case, kind='nonharm', gain=mean,
                                      lower=mean-radius, upper=mean+radius, passed=mean-radius >= -.01))
                    if case in ('copied20', 'mixed20'):
                        gates.append(dict(split=split, case=case, kind='gain', gain=mean,
                                          lower=mean-radius, upper=mean+radius, passed=mean >= .005 and mean-radius > 0))
    assert len(gates) == 14
    return dict(records=records, summaries=summaries, gates=gates,
                screen_passed=all(g['passed'] for g in gates))


if __name__ == '__main__':
    print(json.dumps(run(json.load(sys.stdin))))
