import copy
import hashlib
import json
from pathlib import Path
import unittest

import sparse_v1 as s

ROOT = Path(__file__).parent


class SparseTests(unittest.TestCase):
    def setUp(self):
        self.rows = json.loads((ROOT/'design-preflight-v1.json').read_text())
        self.qs = {q['case_id']: q for q in json.loads((ROOT/'prepared-v2/queries.json').read_text())}
        self.oracle = json.loads((ROOT/'prepared-v2/oracle.json').read_text())

    def test_replay_and_source_hashes(self):
        saved = json.loads((ROOT/'sparse-v1-results.json').read_text())
        for path, digest in saved['sha256'].items():
            self.assertEqual(hashlib.sha256((ROOT/path).read_bytes()).hexdigest(), digest)
        replay = s.evaluate(self.rows, self.qs, self.oracle)
        self.assertEqual(replay['metrics'], saved['metrics'])
        self.assertEqual(replay['predictions'], saved['predictions'])

    def test_held_out_labels_cannot_train_own_predictor(self):
        first = s.evaluate(self.rows, self.qs, self.oracle)
        changed = copy.deepcopy(self.oracle)
        for v in changed.values():
            if v['cluster'] == 'voyager1' and v['support']:
                v['support'] = ['m10-venus']
        second = s.evaluate(self.rows, self.qs, changed)
        for a, b in zip(first['predictions'], second['predictions']):
            if a['held_out'] == 'voyager1':
                self.assertEqual([c['learned'] for c in a['candidates']],
                                 [c['learned'] for c in b['candidates']])
                self.assertTrue(all(self.oracle[q]['cluster'] != 'voyager1'
                                    for q in a['training_queries']))

    def test_boundaries_and_determinism(self):
        frame = self.rows[0]['candidates'][0]['frame']
        x = s.features('Mariner Venus flyby', frame)
        self.assertEqual(x, s.features('Mariner Venus flyby', frame))
        self.assertTrue(all(0 <= k < s.DIM for k in x))
        for query in ('', 'a the', 'x'*1025, ' '.join('word'+str(i) for i in range(65))):
            with self.assertRaises(ValueError):
                s.features(query, frame)
        with self.assertRaises(ValueError):
            s.features('query', 'who: someone')
        with self.assertRaises(ValueError):
            s.features('query', frame+'\nwho: duplicate')
        with self.assertRaises(ValueError):
            s.fit([])


if __name__ == '__main__':
    unittest.main()
