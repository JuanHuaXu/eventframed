import hashlib
import json
from pathlib import Path
import unittest


class NobelAudit(unittest.TestCase):
    def test_hashes_counts_and_metrics(self):
        root=Path(__file__).parent
        saved=json.loads((root/'nobel-v1/summary.json').read_text())
        for p,h in saved['sha256'].items():
            self.assertEqual(hashlib.sha256((root.parent.parent/p).read_bytes()).hexdigest(),h)
        corpus=json.loads((root/'nobel-v1/corpus.json').read_text())
        queries=json.loads((root/'nobel-v1/queries.json').read_text())
        oracle=json.loads((root/'nobel-v1/oracle.json').read_text())
        self.assertEqual(len(corpus),19);self.assertEqual(len(queries),14)
        self.assertEqual(len({r['fixture_id'] for r in corpus}),19)
        for arm in ('baseline','lexical','sparse'):
            rows=json.loads((root/f'nobel-v1/{arm}-results.json').read_text())
            self.assertEqual({r['case'] for r in rows},set(oracle))
            for kind in ('literal','paraphrase','absent'):
                selected=[r for r in rows if oracle[r['case']]['wording']==kind]
                for r in selected:
                    expected=next((i+1 for i,c in enumerate(r['candidates']) if c['fixture'] in oracle[r['case']]['support']),0)
                    self.assertEqual(r['support_rank'],expected)
                m=saved['report'][arm][kind]
                self.assertEqual(m['top1'],sum(r['support_rank']==1 for r in selected))
                self.assertEqual(m['survived'],sum(r['support_rank']>0 for r in selected))


if __name__=='__main__':unittest.main()
