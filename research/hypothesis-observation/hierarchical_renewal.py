"""Exact mixture of shared and type-specific renewal mechanisms."""
import math
import experiment
from uncertain_renewal import UncertainRenewal


class HierarchicalRenewal:
    def __init__(self, table):
        self.models = [UncertainRenewal(table, p) for p in (0, 1, .5)]
        # Half prior mass on a shared mechanism (split equally copy/fresh),
        # half on independent per-type mechanisms. No fitted mixing weight.
        self.logs = [math.log(p) for p in (.25, .25, .5)]
        self.used = [[0] * 8, [0] * 8]
        self.log_evidence = 0.0
        self.w = [1 / 16] * 16

    def weights(self):
        return [math.exp(x) for x in self.logs]

    def row(self, action):
        kind, t, slot = action
        assert kind in (0, 1) and 0 <= t < 8 and slot == self.used[kind][t]
        numerator = [0.0] * 16
        for mass, model in zip(self.weights(), self.models):
            if mass == 0:
                continue
            row = model.row(action)
            for h in range(16):
                numerator[h] += mass * model.w[h] * row[h]
        # A zero-weight h cannot affect the mixture forecast. This convention
        # defines its conditional row without fabricating positive evidence.
        return [n/w if w > 0 else .5 for n, w in zip(numerator, self.w)]

    def observe(self, action, outcome):
        assert type(outcome) is bool
        self.row(action)
        logs = []
        for log_mass, model in zip(self.logs, self.models):
            if log_mass == -math.inf:
                logs.append(-math.inf)
                continue
            row = model.row(action)
            likelihood = sum(w * (p if outcome else 1-p) for w, p in zip(model.w, row))
            if likelihood == 0:
                # Contradiction eliminates this fixed component, not the data.
                # Inactive components are never queried or revived afterward.
                logs.append(-math.inf)
                continue
            assert likelihood > 0
            model.observe(action, outcome)
            logs.append(log_mass + math.log(likelihood))
        maximum = max(logs)
        assert math.isfinite(maximum)
        normalizer = maximum + math.log(sum(math.exp(x-maximum) for x in logs))
        self.logs = [x-normalizer for x in logs]
        self.log_evidence += normalizer
        masses = self.weights()
        self.w = [sum(m * model.w[h] for m, model in zip(masses, self.models)) for h in range(16)]
        self.used[action[0]][action[1]] += 1
        assert abs(sum(self.w)-1) < 1e-12

    def forecast(self):
        return experiment.classes(self.w, 'noise05')

    def fresh_probability(self, t):
        return sum(w * model.fresh_probability(t) for w, model in zip(self.weights(), self.models) if w > 0)
