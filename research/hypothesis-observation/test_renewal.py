import copy
import random
import unittest
import experiment
from renewal import Observer, episode


class RenewalTests(unittest.TestCase):
    def test_fresh_does_not_relabel_ordinary_source(self):
        o = Observer(experiment.likelihood('noise20'))
        o.observe((0, 0, 0), True)
        mode, first, ordinary = copy.deepcopy(o.model.mode), o.model.first[:], o.used[0][:]
        before = o.model.w[:]
        o.observe((1, 0, 0), False)
        self.assertNotEqual(before, o.model.w)
        self.assertEqual(mode, o.model.mode)
        self.assertEqual(first, o.model.first)
        self.assertEqual(ordinary, o.used[0])

    def test_clone_and_selection_do_not_mutate_live_state(self):
        o = Observer(experiment.likelihood('noise20'))
        before = copy.deepcopy(o.__dict__)
        child = o.clone()
        child.observe((0, 1, 0), True)
        o.select(16, 'mixed', random.Random(1))
        self.assertEqual(before['model'].__dict__, o.model.__dict__)
        self.assertEqual(before['used'], o.used)

    def test_one_credit_uses_only_ordinary(self):
        o = Observer(experiment.likelihood('noise20'))
        for arm in ('regular', 'mixed', 'random', 'entropy'):
            self.assertEqual(o.select(1, arm, random.Random(2))[0], 0)

    def test_ordinary_model_matches_original(self):
        from dependence_v3 import Joint
        table = experiment.likelihood('noise20')
        o, original = Observer(table), Joint(table)
        for t, y in [(0, True), (0, False), (1, True), (3, False)]:
            slot = o.used[0][t]
            o.observe((0, t, slot), y)
            original.observe(t, y)
            self.assertEqual(o.model.__dict__, original.__dict__)

    def test_episode_replay_cost_and_prefix(self):
        r = episode('copied20', 771231)
        self.assertEqual(r, episode('copied20', 771231))
        for arm, result in r['arms'].items():
            o = Observer(experiment.likelihood('noise20'))
            # At each prefix, the next action depends only on reconstructed
            # observed state and policy RNG, not the saved future suffix.
            rng = random.Random(771231*10+2)
            for step in result['trace']:
                self.assertEqual(list(o.select(16-step['credit'], arm, rng)), list(step['action']))
                self.assertEqual(o.forecast(), step['forecast'])
                o.observe(step['action'], step['outcome'])
            self.assertEqual(o.forecast(), result['final'])
            self.assertEqual(sum(s['cost'] for s in result['trace']), 16)


if __name__ == '__main__':
    unittest.main()
