import gzip
import hashlib
import json
from pathlib import Path
import unittest
import reliability_v6 as d


class ReliabilityTests(unittest.TestCase):
    def test_joint_enumeration(self):
        table = d.base.old.likelihood('noise20')
        signals = [0, 1, 1, 0, 1, 0, 1, 0]
        matrices = d.likelihoods(d.previous.prior.calibrate())
        model = d.Reliability(table, signals, matrices)
        history = []

        def check():
            masses = [[0.0]*16 for _ in matrices]
            for r, matrix in enumerate(matrices):
                for h in range(16):
                    for modes in range(256):
                        p = 1/(3*16*256)
                        for t, signal in enumerate(signals):
                            p *= matrix[int(bool(modes & (1 << t)))][signal]
                        first = {}
                        for t, y in history:
                            if modes & (1 << t) or t not in first:
                                p *= table[t][h] if y else 1-table[t][h]
                            elif first[t] != y:
                                p = 0
                            first.setdefault(t, y)
                        masses[r][h] += p
            total = sum(map(sum, masses))
            for r in range(3):
                self.assertAlmostEqual(model.weights[r], sum(masses[r])/total, places=12)
            for h in range(16):
                self.assertAlmostEqual(model.hypothesis()[h], sum(row[h] for row in masses)/total, places=12)

        check()
        for t, y in [(0, True), (1, False), (0, True), (1, True), (7, False)]:
            history.append((t, y))
            model.observe(t, y)
            check()

    def test_first_reports_cannot_identify_copying(self):
        m = d.Reliability(d.base.old.likelihood('noise20'), [0, 1]*4,
                          d.likelihoods(d.previous.prior.calibrate()))
        initial = m.weights[:]
        for t in range(8):
            m.observe(t, t % 2 == 0)
            for a, b in zip(initial, m.weights):
                self.assertAlmostEqual(a, b, places=12)

    def test_replay(self):
        root = Path(__file__).parent
        path = root/'reliability-v6.json.gz'
        if not path.exists():
            self.skipTest('experiment not yet run')
        with gzip.open(path, 'rt') as f:
            saved = json.load(f)
        self.assertEqual(saved['calibration'], d.previous.prior.calibrate())
        for p, digest in saved['hashes'].items():
            self.assertEqual(hashlib.sha256((root/p).read_bytes()).hexdigest(), digest)
        for r in saved['records']:
            replay = d.episode(r['case'], r['seed'], saved['calibration'])
            replay['split'] = r['split']
            self.assertEqual(json.loads(json.dumps(replay)), r)
            traces = [r['arms'][a]['trace'] for a in d.ARMS]
            for steps in zip(*traces):
                self.assertTrue(all(x['pair'] == steps[0]['pair'] and x['outcome'] == steps[0]['outcome'] for x in steps))
                for x in steps:
                    self.assertAlmostEqual(sum(x['forecast']), 1)
                    self.assertTrue(all(0 <= p <= 1 for p in x['forecast']))
                self.assertAlmostEqual(sum(steps[2]['mode_weights']), 1)
            self.assertEqual(len({tuple(x['pair']) for x in traces[0]}), 16)
            self.assertAlmostEqual(sum(r['arms']['average']['mode_weights']), 1)
        for k, v in d.summarize(saved['records']).items():
            self.assertEqual(saved[k], v)


if __name__ == '__main__':
    unittest.main()
