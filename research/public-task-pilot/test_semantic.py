import hashlib
import json
from pathlib import Path
import unittest


class SemanticAudit(unittest.TestCase):
    def test_saved_evidence(self):
        root=Path(__file__).parent
        out=json.loads((root/'nobel-v1/semantic-results.json').read_text())
        oracle=json.loads((root/'nobel-v1/oracle.json').read_text())
        for p,h in out['Hashes'].items():
            self.assertEqual(hashlib.sha256((root.parent.parent/p).read_bytes()).hexdigest(),h)
        self.assertEqual(len(out['Results']),42)
        for arm in ('baseline','lexical','sparse'):
            rows=[r for r in out['Results'] if r['Arm']==arm]
            self.assertEqual({r['Case'] for r in rows},set(oracle))
            for r in rows:
                self.assertGreater(r['RecallNS'],0)
                if arm!='baseline':self.assertEqual(r['Frontier'],19)
                expected=next((i+1 for i,c in enumerate(r['Candidates']) if c['Fixture'] in oracle[r['Case']]['support']),0)
                self.assertEqual(expected,r['SupportRank'])
                for c in r['Candidates']:
                    self.assertAlmostEqual(c['Law']['useful']+c['Law']['not_useful'],1)
            literal=[r for r in rows if oracle[r['Case']]['wording']=='literal']
            para=[r for r in rows if oracle[r['Case']]['wording']=='paraphrase']
            self.assertEqual(sum(r['SupportRank']==1 for r in literal),6)
            self.assertEqual(sum(r['SupportRank']==1 for r in para),{'baseline':5,'lexical':3,'sparse':2}[arm])


if __name__=='__main__':unittest.main()
