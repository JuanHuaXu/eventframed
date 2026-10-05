import gzip
import hashlib
import json
from pathlib import Path
import random
import unittest
from unittest.mock import patch
import lookahead_rollout_v9 as d


class RolloutTests(unittest.TestCase):
    def test_last_step_is_one_step(self):
        with patch.object(d.previous, 'select', return_value=(7, 3)) as one:
            with patch.object(d.planner, 'inspect', side_effect=AssertionError('extra horizon')):
                self.assertEqual(d.action(None, set(), 1), (7, 3))
                one.assert_called_once()

    def test_replay(self):
        root = Path(__file__).parent
        path = root/'lookahead-rollout-v9.json.gz'
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
            rng = random.Random(r['seed']*10+1)
            tape = [[rng.random() for _ in range(8)] for _ in range(4)]
            table = d.previous.base.old.likelihood('noise05' if r['case'].endswith('05') else 'noise20')
            for arm in d.ARMS:
                seen = set()
                for step in r['arms'][arm]['trace']:
                    t, s = step['pair']
                    self.assertNotIn((t, s), seen)
                    seen.add((t, s))
                    self.assertEqual(step['outcome'], tape[s if r['modes'][t] else 0][t] < table[t][r['truth']])
                    self.assertAlmostEqual(sum(step['forecast']), 1)
                    self.assertTrue(all(0 <= p <= 1 for p in step['forecast']))
                self.assertEqual(len(seen), 16)
                self.assertEqual(r['arms'][arm]['cost'], 24)
        for k, v in d.summarize(saved['records']).items():
            self.assertEqual(saved[k], v)


if __name__ == '__main__':
    unittest.main()
