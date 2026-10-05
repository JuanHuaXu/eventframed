import itertools
import math
import unittest
import experiment
from hierarchical_renewal import HierarchicalRenewal


def batch(table, history):
    components = []
    for fresh_prior, prior in [(0, .25), (1, .25), (.5, .5)]:
        weights = []
        for h in range(16):
            w = prior / 16
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
        components.append(weights)
    z = sum(map(sum, components))
    return [sum(c[h] for c in components)/z for h in range(16)], [sum(c)/z for c in components], math.log(z)


class HierarchicalTests(unittest.TestCase):
    def test_all_outcomes_and_channel_orders(self):
        table = experiment.likelihood('noise20')
        paths = [((1, 0, 0), (0, 0, 0), (1, 1, 0), (0, 1, 0), (1, 1, 1), (0, 0, 1)),
                 ((0, 0, 0), (0, 1, 0), (1, 1, 0), (1, 0, 0), (0, 0, 1), (1, 1, 1))]
        for path in paths:
            for outcomes in itertools.product((False, True), repeat=6):
                model, history = HierarchicalRenewal(table), []
                for action, outcome in zip(path, outcomes):
                    predictive = sum(w*p for w, p in zip(model.w, model.row(action)))
                    previous = model.log_evidence
                    model.observe(action, outcome)
                    self.assertAlmostEqual(model.log_evidence-previous,
                                           math.log(predictive if outcome else 1-predictive), places=11)
                    history.append((action, outcome))
                    weights, components, evidence = batch(table, history)
                    for a, b in zip(model.w, weights):
                        self.assertAlmostEqual(a, b, places=11)
                    for a, b in zip(model.weights(), components):
                        self.assertAlmostEqual(a, b, places=11)
                    self.assertAlmostEqual(model.log_evidence, evidence, places=11)

    def test_mixed_modes_do_not_force_one_global_truth(self):
        model = HierarchicalRenewal([[.5] * 16 for _ in range(8)])
        model.observe((0, 0, 0), False)
        model.observe((1, 0, 0), True)
        self.assertEqual(model.weights()[0], 0)
        self.assertAlmostEqual(model.fresh_probability(0), 1, places=12)
        self.assertLess(model.fresh_probability(1), 1)
        model.observe((0, 1, 0), True)
        for i in range(8):
            model.observe((1, 1, i), True)
        self.assertGreater(model.weights()[2], .9)
        self.assertLess(model.fresh_probability(1), .05)
        self.assertAlmostEqual(model.fresh_probability(0), 1, places=12)

    def test_no_cross_type_information_without_renewal(self):
        model = HierarchicalRenewal(experiment.likelihood('noise20'))
        for i, y in enumerate([True, False, True, False]):
            model.observe((0, 0, i), y)
        for actual, prior in zip(model.weights(), [.25, .25, .5]):
            self.assertAlmostEqual(actual, prior, places=12)
        model.weights().clear()
        self.assertEqual(len(model.weights()), 3)


if __name__ == '__main__':
    unittest.main()
