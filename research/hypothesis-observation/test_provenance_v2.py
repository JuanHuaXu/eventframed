import gzip
import hashlib
import json
from pathlib import Path
import random
import unittest
import provenance_v2 as p


class ProvenanceTests(unittest.TestCase):
    def test_source_exhaustion_and_coin(self):
        table=p.old.likelihood('noise05')
        self.assertIsNone(p.choose([1/16]*16,table,{(t,0) for t in range(8)},1,'source_gini',random.Random(1)))
        w=[1/16]*16
        self.assertEqual(p.old.update(w,[.5]*16,True),w)
        r=p.episode('one_source',123)
        for arm in p.ARMS[1:]:
            trace=r['arms'][arm]['trace']
            pairs=[tuple(x['pair']) for x in trace if x['pair'] is not None]
            self.assertEqual(len(pairs),8)
            self.assertEqual(len(set(pairs)),8)
            for x in trace[8:]:
                self.assertIsNone(x['outcome'])
                self.assertEqual(x['forecast'],r['arms'][arm]['final'])

    def test_full_replay_and_journal(self):
        root=Path(__file__).parent
        with gzip.open(root/'provenance-v2.json.gz','rt') as f:
            saved=json.load(f)
        for path,h in saved['hashes'].items():
            self.assertEqual(hashlib.sha256((root/path).read_bytes()).hexdigest(),h)
        for r in saved['records']:
            replay=p.episode(r['case'],r['seed'])
            replay['split']=r['split']
            self.assertEqual(json.loads(json.dumps(replay)),r)
            table=p.old.likelihood('noise20' if r['case']=='noise20' else 'noise05')
            for arm in p.ARMS:
                w=[1/16]*16
                seen=set()
                for step in r['arms'][arm]['trace']:
                    self.assertEqual(step['forecast'],p.old.classes(w,'noise05'))
                    if step['pair'] is not None:
                        pair=tuple(step['pair'])
                        if arm!='naive':self.assertNotIn(pair,seen)
                        seen.add(pair)
                        w=p.old.update(w,table[pair[0]],step['outcome'])
                self.assertEqual(p.old.classes(w,'noise05'),r['arms'][arm]['final'])
        summary=p.summarize(saved['records'])
        for k,v in summary.items():self.assertEqual(v,saved[k])


if __name__=='__main__':unittest.main()
