import copy
import gzip
import hashlib
import json
from pathlib import Path
import random
import unittest
import mixture_acquisition_v7 as d


class AcquisitionTests(unittest.TestCase):
    def test_gain_counterfactuals_and_no_mutation(self):
        calibration = d.previous.previous.prior.calibrate()
        model = d.previous.Reliability(d.base.old.likelihood('noise20'), [0, 1]*4,
                                       d.previous.likelihoods(calibration))
        for t, y in [(0, True), (1, False), (0, True), (1, True)]:
            model.observe(t, y)
        before = copy.deepcopy(model.__dict__)
        for t in range(8):
            p = sum(w*sum(h*q for h, q in zip(m.w, m.row(t)))
                    for w, m in zip(model.weights, model.models))
            expected = -sum(x*x for x in model.forecast())
            for y, probability in [(False, 1-p), (True, p)]:
                after = copy.deepcopy(model)
                after.observe(t, y)
                expected += probability*sum(x*x for x in after.forecast())
            self.assertAlmostEqual(d.gain(model, t), expected, places=12)
            self.assertGreaterEqual(d.gain(model, t), -1e-12)
        chosen = d.select(model, set())
        self.assertIsNotNone(chosen)
        self.assertIsNone(d.select(model, {(t, s) for t in range(8) for s in range(4)}))
        self.assertEqual(model.weights, before['weights'])
        for m, b in zip(model.models, before['models']):
            self.assertEqual(m.__dict__, b.__dict__)

    def test_replay(self):
        root = Path(__file__).parent
        path = root/'mixture-acquisition-v7.json.gz'
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
            table = d.base.old.likelihood('noise05' if r['case'].endswith('05') else 'noise20')
            for arm in d.ARMS:
                seen = set()
                for step in r['arms'][arm]['trace']:
                    t, s = step['pair']
                    self.assertNotIn((t, s), seen)
                    seen.add((t, s))
                    self.assertEqual(step['outcome'], tape[s if r['modes'][t] else 0][t] < table[t][r['truth']])
                    self.assertAlmostEqual(sum(step['forecast']), 1)
                self.assertEqual(len(seen), 16)
                self.assertEqual(r['arms'][arm]['cost'], 24)
        for k, v in d.summarize(saved['records']).items():
            self.assertEqual(saved[k], v)


if __name__ == '__main__':
    unittest.main()
