import itertools
import unittest
import experiment
from renewal import Observer
from uncertain_renewal import UncertainRenewal


def batch(table, history, fresh_prior=.5):
    weights = []
    for h in range(16):
        w = 1 / 16
        for t in range(8):
            q, mass = table[t][h], 0.0
            for ordinary, fresh, root in itertools.product((0, 1), repeat=3):
                p = .5 * (fresh_prior if fresh else 1-fresh_prior) * (q if root else 1-q)
                for (kind, typ, slot), y in history:
                    if typ != t:
                        continue
                    emission = q if (fresh if kind else ordinary and slot > 0) else root
                    p *= emission if y else 1-emission
                mass += p
            w *= mass
        weights.append(w)
    total = sum(weights)
    return [w / total for w in weights]


class UncertainRenewalTests(unittest.TestCase):
    def test_batch_marginalization_all_six_outcome_sequences(self):
        table = experiment.likelihood('noise20')
        paths = [((0, 0, 0), (1, 0, 0), (0, 0, 1), (1, 0, 1), (0, 1, 0), (1, 1, 0)),
                 ((1, 0, 0), (1, 0, 1), (0, 0, 0), (0, 0, 1), (1, 1, 0), (0, 1, 0))]
        for path in paths:
            for outcomes in itertools.product((False, True), repeat=6):
                model, history = UncertainRenewal(table), []
                for action, y in zip(path, outcomes):
                    model.observe(action, y)
                    history.append((action, y))
                    for a, b in zip(model.w, batch(table, history)):
                        self.assertAlmostEqual(a, b, places=12)
                    for rows in model.local:
                        for row in rows:
                            self.assertAlmostEqual(sum(row), 1, places=12)

    def test_certain_fresh_recovers_original_model(self):
        table = experiment.likelihood('noise20')
        for outcomes in itertools.product((False, True), repeat=4):
            model, original = UncertainRenewal(table, 1), Observer(table)
            for action, y in zip([(1, 0, 0), (0, 0, 0), (0, 0, 1), (1, 0, 1)], outcomes):
                model.observe(action, y)
                original.observe(action, y)
                for a, b in zip(model.w, original.model.w):
                    self.assertAlmostEqual(a, b, places=12)

    def test_disagreement_rules_out_copied_renewal(self):
        model = UncertainRenewal(experiment.likelihood('noise20'))
        model.observe((1, 0, 0), True)
        model.observe((0, 0, 0), False)
        self.assertAlmostEqual(model.fresh_probability(0), 1, places=12)
        self.assertNotAlmostEqual(model.fresh_probability(1), 1, places=12)

    def test_duplicate_rejected(self):
        model = UncertainRenewal(experiment.likelihood('noise20'))
        model.observe((1, 0, 0), True)
        with self.assertRaises(AssertionError):
            model.observe((1, 0, 0), True)


if __name__ == '__main__':
    unittest.main()
