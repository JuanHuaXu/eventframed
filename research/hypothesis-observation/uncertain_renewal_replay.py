"""Re-score original acquired traces; never claim closed-loop equivalence."""
import json
import math
import statistics
import sys
import experiment
from uncertain_renewal import UncertainRenewal


def run(data):
    records = []
    for r in data['records']:
        table = experiment.likelihood('noise05' if r['case'].endswith('05') else 'noise20')
        arms = {}
        for arm, old in r['arms'].items():
            model, forecasts, area = UncertainRenewal(table), [], 0.0
            for step in old['trace']:
                p = model.forecast()
                forecasts.append(p)
                area += step['cost'] * experiment.brier(p, r['truth'] % 4) / 16
                model.observe(step['action'], step['outcome'])
            final = model.forecast()
            chosen = max(range(4), key=final.__getitem__)
            arms[arm] = dict(forecasts=forecasts, final=final,
                final_brier=experiment.brier(final, r['truth'] % 4), credit_brier=area,
                correct=chosen == r['truth'] % 4,
                confident_wrong=max(final) >= .9 and chosen != r['truth'] % 4,
                freshness=[model.fresh_probability(t) for t in range(8)],
                old_final_brier=old['final_brier'], old_confident_wrong=old['confident_wrong'])
        records.append(dict(split=r['split'], case=r['case'], seed=r['seed'], arms=arms))
    summaries, gates = [], []
    for split in range(2):
        for case in ['independent20', 'copied20', 'mixed20', 'copied05', 'false_renewal20']:
            rs = [r for r in records if r['split'] == split and r['case'] == case]
            assert len(rs) == 64
            for arm in ['regular', 'mixed', 'random', 'entropy']:
                ds = [r['arms'][arm]['old_final_brier'] - r['arms'][arm]['final_brier'] for r in rs]
                mean, radius = statistics.mean(ds), 3.3 * statistics.stdev(ds) / math.sqrt(64)
                summaries.append(dict(split=split, case=case, arm=arm, gain=mean,
                    lower=mean-radius, upper=mean+radius,
                    **{m: statistics.mean(r['arms'][arm][m] for r in rs)
                       for m in ['final_brier', 'credit_brier', 'confident_wrong', 'old_confident_wrong', 'correct']}))
                if arm == 'mixed':
                    passed = mean >= .02 and mean-radius > 0 if case == 'false_renewal20' else mean-radius >= -.01
                    gates.append(dict(split=split, case=case, gain=mean, lower=mean-radius,
                                      upper=mean+radius, passed=passed))
    return dict(records=records, summaries=summaries, gates=gates,
                screen_passed=all(g['passed'] for g in gates))


if __name__ == '__main__':
    print(json.dumps(run(json.load(sys.stdin))))
