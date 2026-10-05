import copy
import gzip
import hashlib
import json
from pathlib import Path
import unittest
import stopping_v10 as d


class StoppingTests(unittest.TestCase):
    def test_gate_and_horizon(self):
        self.assertTrue(d.stop_now(.001, .02, 4, False))
        self.assertFalse(d.stop_now(.001, .02, 4, True))
        self.assertTrue(d.stop_now(.005, .01, 4, True))
        self.assertTrue(d.stop_now(.005, None, 1, True))
        self.assertFalse(d.stop_now(.006, None, 1, True))

    def test_suffix_cannot_change_stopped_arm(self):
        trace = [dict(forecast=[.25]*4, pair=(i % 8, i//8), outcome=True) for i in range(16)]
        trace[3]['forecast'] = [.7, .1, .1, .1]
        original = d.arm(trace, [.1, .7, .1, .1], 0, 3)
        changed = copy.deepcopy(trace)
        changed[3]['outcome'] = False
        changed[3]['pair'] = (7, 3)
        for i in range(4, 16):
            changed[i] = {}
        self.assertEqual(d.arm(changed, [0, 0, 0, 1], 0, 3), original)
        self.assertEqual(original['cost'], 11)

    def test_full_policy_parity(self):
        calibration = d.base.previous.previous.prior.calibrate()
        for case in ['independent20', 'mixed20', 'matched_misleading20']:
            a = d.episode(case, 517251, calibration)['arms']['lookahead']
            b = d.previous.episode(case, 517251, calibration)['arms']['lookahead']
            for field in ['final', 'curve_brier', 'final_brier', 'correct', 'confident_wrong', 'cost']:
                self.assertEqual(a[field], b[field])
            for x, y in zip(a['trace'], b['trace']):
                for field in ['pair', 'outcome', 'forecast']:
                    self.assertEqual(x[field], y[field])

    def test_replay(self):
        root = Path(__file__).parent
        path = root/'stopping-v10.json.gz'
        if not path.exists():
            self.skipTest('experiment not yet run')
        with gzip.open(path, 'rt') as f:
            saved = json.load(f)
        for p, digest in saved['hashes'].items():
            self.assertEqual(hashlib.sha256((root/p).read_bytes()).hexdigest(), digest)
        for r in saved['records']:
            replay = d.episode(r['case'], r['seed'], saved['calibration'])
            replay['split'] = r['split']
            self.assertEqual(json.loads(json.dumps(replay)), r)
            for name, cautious in [('stop_one', False), ('stop_two', True)]:
                a = r['arms'][name]
                stop = a['stop_at']
                seen = set()
                for i, x in enumerate(a['trace']):
                    if i < stop:
                        pair = tuple(x['pair'])
                        self.assertNotIn(pair, seen)
                        seen.add(pair)
                        gate = r['decisions'][i]
                        self.assertFalse(d.stop_now(gate['one'], gate['two'], 16-i, cautious))
                    else:
                        self.assertIsNone(x['pair'])
                        self.assertIsNone(x['outcome'])
                        self.assertEqual(x['forecast'], a['final'])
                    self.assertAlmostEqual(sum(x['forecast']), 1)
                if stop < 16:
                    gate = r['decisions'][stop]
                    self.assertTrue(d.stop_now(gate['one'], gate['two'], 16-stop, cautious))
                self.assertEqual(a['reports'], len(seen))
                self.assertEqual(a['cost'], 8+len(seen))
                self.assertAlmostEqual(a['penalized'], a['final_brier']+d.PRICE*a['cost'])
        for k, v in d.summarize(saved['records']).items():
            self.assertEqual(saved[k], v)


if __name__ == '__main__':
    unittest.main()
