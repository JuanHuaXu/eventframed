import copy
import random
import unittest
import experiment
import renewal
from adaptive_renewal import AdaptiveObserver, episode, policy


class AdaptiveTests(unittest.TestCase):
    def test_branch_ownership(self):
        o = AdaptiveObserver(experiment.likelihood('noise20'))
        original = copy.deepcopy(o.model.__dict__)
        a, b = o.clone(), o.clone()
        a.observe((1, 0, 0), True)
        b.observe((0, 0, 0), False)
        a.observe((0, 0, 0), True)
        self.assertEqual(original, o.model.__dict__)
        self.assertNotEqual(a.model.w, b.model.w)
        self.assertIs(a.used, a.model.used)
        o.select(16, 'mixed', random.Random(1))
        self.assertEqual(original, o.model.__dict__)

    def test_control_parity_and_all_uncertain_action_prefixes(self):
        seed = 112671
        r, old = episode('mixed20', seed), renewal.episode('mixed20', seed)
        self.assertEqual(r['arms']['regular'], old['arms']['regular'])
        self.assertEqual(r['arms']['certain_mixed'], old['arms']['mixed'])
        for arm, a in r['arms'].items():
            if not arm.startswith('uncertain_'):
                continue
            o, rng = AdaptiveObserver(experiment.likelihood('noise20')), random.Random(seed*10+2)
            for s in a['trace']:
                self.assertEqual(o.select(16-s['credit'], policy(arm), rng), s['action'])
                self.assertEqual(o.forecast(), s['forecast'])
                o.observe(s['action'], s['outcome'])
            self.assertEqual(o.forecast(), a['final'])
            self.assertEqual(sum(s['cost'] for s in a['trace']), 16)


if __name__ == '__main__':
    unittest.main()
