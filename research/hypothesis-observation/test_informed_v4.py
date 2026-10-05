import gzip
import hashlib
import json
from pathlib import Path
import unittest
import informed_v4 as d


class InformedTests(unittest.TestCase):
    def test_unequal_prior_enumeration(self):
        priors = [.1, .9, .2, .8, .3, .7, .4, .6]
        model = d.base.Joint(d.base.old.likelihood('noise20'))
        model.mode = [[p]*16 for p in priors]
        history = []
        for t, y in [(0, True), (1, False), (0, True), (1, True), (7, False)]:
            model.observe(t, y)
            history.append((t, y))
            joint = []
            for h in range(16):
                for modes in range(256):
                    p = 1/16
                    for k, prior in enumerate(priors):
                        p *= prior if modes & (1 << k) else 1-prior
                    first = {}
                    for test, outcome in history:
                        if modes & (1 << test) or test not in first:
                            p *= model.table[test][h] if outcome else 1-model.table[test][h]
                        elif first[test] != outcome:
                            p = 0
                        first.setdefault(test, outcome)
                    joint.append((h, modes, p))
            total = sum(p for _, _, p in joint)
            for h in range(16):
                mass = sum(p for hh, _, p in joint if hh == h)
                self.assertAlmostEqual(model.w[h], mass/total, places=12)
                for k in range(8):
                    independent = sum(p for hh, m, p in joint if hh == h and m & (1 << k))
                    self.assertAlmostEqual(model.mode[k][h], independent/mass, places=12)

    def test_calibration_isolation(self):
        c = d.calibrate()
        self.assertEqual(sum(map(sum, c['counts'])), 4096)
        self.assertLess(c['priors'][0], .5)
        self.assertGreater(c['priors'][1], .5)
        original = d.episode('mixed20', 81234, c)
        altered = d.episode('mixed20', 81234, dict(priors=[.5, .5]))
        for arm in ['joint_gini', 'independent', 'once']:
            self.assertEqual(original['arms'][arm], altered['arms'][arm])
        self.assertEqual(c, d.calibrate())

    def test_replay(self):
        root = Path(__file__).parent
        path = root/'informed-v4.json.gz'
        if not path.exists():
            self.skipTest('experiment artifact not yet generated')
        with gzip.open(path, 'rt') as f:
            saved = json.load(f)
        self.assertEqual(saved['calibration'], d.calibrate())
        for p, digest in saved['hashes'].items():
            self.assertEqual(hashlib.sha256((root/p).read_bytes()).hexdigest(), digest)
        for record in saved['records']:
            replay = d.episode(record['case'], record['seed'], saved['calibration'])
            replay['split'] = record['split']
            self.assertEqual(json.loads(json.dumps(replay)), record)
            for arm in d.ARMS:
                seen = set()
                for step in record['arms'][arm]['trace']:
                    self.assertAlmostEqual(sum(step['forecast']), 1)
                    if step['pair'] is not None:
                        pair = tuple(step['pair'])
                        self.assertNotIn(pair, seen)
                        seen.add(pair)
                self.assertEqual(record['arms'][arm]['cost'], 8+len(seen))
        for k, value in d.summarize(saved['records']).items():
            self.assertEqual(saved[k], value)


if __name__ == '__main__':
    unittest.main()
