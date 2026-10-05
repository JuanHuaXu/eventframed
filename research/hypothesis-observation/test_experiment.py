import math
import random
import unittest
import experiment as e


class Tests(unittest.TestCase):
    def test_coin_does_not_update(self):
        w = [1/16]*16
        self.assertEqual(w, e.update(w, [.5]*16, True))
        self.assertEqual(w, e.update(w, [.5]*16, False))

    def test_target_avoids_nuisance(self):
        table = e.likelihood("noise05")
        self.assertIn(e.select([1/16]*16, table, "noise05", "target_gini", random.Random(1)), [0,1])
        self.assertIn(e.select([1/16]*16, table, "noise05", "full_information", random.Random(1)), [2,3])

    def test_journal(self):
        for case in e.CASES:
            rec = e.episode(case, 2026100301)
            for arm in e.ARMS:
                w = [1/16]*16
                rng = random.Random(rec["seed"]*10+2)
                for tick in rec["arms"][arm]["trace"]:
                    self.assertEqual(tick["forecast"], e.classes(w, case))
                    self.assertEqual(tick["test"], e.select(w, e.likelihood(case), case, arm, rng))
                    w = e.update(w, e.likelihood(case)[tick["test"]], tick["outcome"])
                    self.assertTrue(math.isclose(sum(w), 1))
                self.assertEqual(e.classes(w, case), rec["arms"][arm]["final"])


if __name__ == "__main__":
    unittest.main()
