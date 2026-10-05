"""Exact local latent-mode model for renewal evidence of uncertain origin."""
import experiment

STATES = [(ordinary, fresh, root) for ordinary in (0, 1)
          for fresh in (0, 1) for root in (0, 1)]


class UncertainRenewal:
    def __init__(self, table, fresh_prior=.5):
        assert 0 <= fresh_prior <= 1
        self.table = table
        self.w = [1 / 16] * 16
        self.used = [[0] * 8, [0] * 8]
        # Per-type local state is conditional on h; the shared root exists
        # before either path is queried, including renewal-before-original.
        self.local = [[[.5 * (fresh_prior if f else 1-fresh_prior)
                        * (q if root else 1-q) for _, f, root in STATES]
                       for q in row] for row in table]

    def emission(self, action, h, state):
        kind, t, slot = action
        ordinary, fresh, root = STATES[state]
        independent = fresh if kind else ordinary and slot > 0
        return self.table[t][h] if independent else float(root)

    def row(self, action):
        kind, t, slot = action
        assert kind in (0, 1) and 0 <= t < 8 and slot == self.used[kind][t]
        return [sum(m * self.emission(action, h, s) for s, m in enumerate(local))
                for h, local in enumerate(self.local[t])]

    def observe(self, action, outcome):
        assert type(outcome) is bool
        kind, t, _ = action
        row = self.row(action)
        updated = experiment.update(self.w, row, outcome)
        for h in range(16):
            likelihood = row[h] if outcome else 1-row[h]
            assert likelihood > 0
            self.local[t][h] = [m * (self.emission(action, h, s) if outcome
                                    else 1-self.emission(action, h, s)) / likelihood
                                for s, m in enumerate(self.local[t][h])]
        self.w = updated
        self.used[kind][t] += 1

    def forecast(self):
        return experiment.classes(self.w, 'noise05')

    def fresh_probability(self, t):
        return sum(w * sum(m for s, m in enumerate(self.local[t][h]) if STATES[s][1])
                   for h, w in enumerate(self.w))
