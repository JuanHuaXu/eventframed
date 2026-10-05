import gzip
import hashlib
import json
from pathlib import Path
import unittest
import factorial_v5 as d


class FactorialTests(unittest.TestCase):
    def test_diagonal_parity_and_isolation(self):
        calibration = d.prior.calibrate()
        for case in ['independent20', 'copied20', 'mixed20']:
            for seed in [81234, 719435]:
                r = d.episode(case, seed, calibration)
                old = d.prior.episode(case, seed, calibration)
                for newarm, oldarm in [('FF', 'joint_gini'), ('II', 'informed')]:
                    for field in ['trace', 'final', 'curve_brier', 'final_brier', 'correct', 'confident_wrong', 'cost']:
                        self.assertEqual(r['arms'][newarm][field], old['arms'][oldarm][field])
                changed = d.episode(case, seed, dict(priors=[.5, .5]))
                self.assertEqual(r['arms']['FF'], changed['arms']['FF'])
                self.assertEqual(changed['arms']['FF'], changed['arms']['II'])

    def test_replay(self):
        root = Path(__file__).parent
        path = root/'factorial-v5.json.gz'
        if not path.exists():
            self.skipTest('experiment not yet run')
        with gzip.open(path, 'rt') as f:
            saved = json.load(f)
        self.assertEqual(saved['calibration'], d.prior.calibrate())
        for p, digest in saved['hashes'].items():
            self.assertEqual(hashlib.sha256((root/p).read_bytes()).hexdigest(), digest)
        for r in saved['records']:
            replay = d.episode(r['case'], r['seed'], saved['calibration'])
            replay['split'] = r['split']
            self.assertEqual(json.loads(json.dumps(replay)), r)
            for acquisition in ['F', 'I']:
                first = r['arms'][acquisition+'F']
                second = r['arms'][acquisition+'I']
                seen = set()
                for a, b in zip(first['trace'], second['trace']):
                    self.assertEqual(a['pair'], b['pair'])
                    self.assertEqual(a['outcome'], b['outcome'])
                    pair = tuple(a['pair'])
                    self.assertNotIn(pair, seen)
                    seen.add(pair)
                    for x in [a, b]:
                        self.assertAlmostEqual(sum(x['forecast']), 1)
                        self.assertTrue(all(0 <= v <= 1 for v in x['forecast']))
                self.assertEqual(len(seen), 16)
                self.assertEqual(first['cost'], 24)
                self.assertEqual(second['cost'], 24)
        for k, v in d.summarize(saved['records']).items():
            self.assertEqual(saved[k], v)


if __name__ == '__main__':
    unittest.main()
