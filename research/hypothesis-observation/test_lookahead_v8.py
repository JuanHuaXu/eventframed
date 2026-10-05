import copy
import gzip
import hashlib
import json
from pathlib import Path
import unittest
import lookahead_v8 as d


class LookaheadTests(unittest.TestCase):
    def model(self):
        calibration = d.previous.previous.previous.prior.calibrate()
        return d.previous.previous.Reliability(d.previous.base.old.likelihood('noise20'), [0, 1]*4,
                                               d.previous.previous.likelihoods(calibration))

    def test_four_leaf_values(self):
        model = self.model()
        model.observe(0, True)
        model.observe(0, True)
        seen = {(0, 0), (0, 1)}
        before = copy.deepcopy(model)
        norm = lambda m: sum(p*p for p in m.forecast())
        for score in d.values(model, seen):
            t, _ = score['pair']
            p = d.probability(model, t)
            terminal = 0.0
            for y, chance in [(False, 1-p), (True, p)]:
                first = copy.deepcopy(model)
                first.observe(t, y)
                alternatives = []
                for second, _ in d.candidates(seen | {score['pair']}):
                    q = d.probability(first, second)
                    expected = 0.0
                    for z, probability in [(False, 1-q), (True, q)]:
                        leaf = copy.deepcopy(first)
                        leaf.observe(second, z)
                        expected += probability*norm(leaf)
                    alternatives.append(expected)
                terminal += chance*max(alternatives)
            self.assertAlmostEqual(score['two'], terminal-norm(model), places=12)
        self.assertEqual(model.weights, before.weights)
        for a, b in zip(model.models, before.models):
            self.assertEqual(a.__dict__, b.__dict__)

    def test_boundaries(self):
        model = self.model()
        allpairs = {(t, s) for t in range(8) for s in range(4)}
        self.assertEqual(d.inspect(model, allpairs)['scores'], [])
        one = d.values(model, allpairs-{(7, 3)})
        self.assertEqual(len(one), 1)
        self.assertAlmostEqual(one[0]['one'], one[0]['two'])

    def test_replay(self):
        root = Path(__file__).parent
        if not (root/'lookahead-v8.json').exists():
            self.skipTest('diagnostic not yet run')
        saved = json.loads((root/'lookahead-v8.json').read_text())
        parentpath = root/'mixture-acquisition-v7.json.gz'
        self.assertEqual(hashlib.sha256(parentpath.read_bytes()).hexdigest(), saved['parent_hash'])
        for p, digest in saved['hashes'].items():
            self.assertEqual(hashlib.sha256((root/p).read_bytes()).hexdigest(), digest)
        with gzip.open(parentpath, 'rt') as f:
            parent = json.load(f)
        records = d.run_records(parent)
        self.assertEqual(json.loads(json.dumps(records)), saved['records'])
        self.assertEqual(len(records), 448)
        for r in records:
            self.assertGreaterEqual(r['value_gain'], -1e-12)
        for k, v in d.summarize(records).items():
            self.assertEqual(saved[k], v)


if __name__ == '__main__':
    unittest.main()
