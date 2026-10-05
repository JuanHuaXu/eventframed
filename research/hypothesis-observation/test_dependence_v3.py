import gzip
import hashlib
import json
from pathlib import Path
import unittest
import dependence_v3 as d


class DependenceTests(unittest.TestCase):
    def test_exact_enumeration(self):
        table=d.old.likelihood('noise20')
        m=d.Joint(table)
        history=[]
        for t,y in [(0,True),(1,False),(0,True),(1,True),(7,False)]:
            m.observe(t,y);history.append((t,y))
            joint=[]
            for h in range(16):
                for modes in range(256):
                    p=1/4096;first={}
                    for test,outcome in history:
                        if modes & (1<<test) or test not in first:
                            p*=table[test][h] if outcome else 1-table[test][h]
                        elif first[test]!=outcome:p=0
                        first.setdefault(test,outcome)
                    joint.append((h,modes,p))
            total=sum(x[2] for x in joint)
            for h in range(16):
                mass=sum(p for hh,_,p in joint if hh==h)
                self.assertAlmostEqual(m.w[h],mass/total,places=12)
                for test in range(8):
                    marginal=sum(p for hh,mode,p in joint if hh==h and mode&(1<<test))
                    self.assertAlmostEqual(m.mode[test][h],marginal/mass,places=12)
        self.assertTrue(all(abs(v-1)<1e-12 for v in m.mode[1]))

    def test_replay(self):
        root=Path(__file__).parent
        with gzip.open(root/'dependence-v3.json.gz','rt') as f:saved=json.load(f)
        for p,h in saved['hashes'].items():self.assertEqual(hashlib.sha256((root/p).read_bytes()).hexdigest(),h)
        for r in saved['records']:
            replay=d.episode(r['case'],r['seed']);replay['split']=r['split']
            self.assertEqual(json.loads(json.dumps(replay)),r)
            for arm in d.ARMS:
                seen=set()
                for step in r['arms'][arm]['trace']:
                    self.assertAlmostEqual(sum(step['forecast']),1)
                    if step['pair'] is not None:
                        pair=tuple(step['pair']);self.assertNotIn(pair,seen);seen.add(pair)
                self.assertTrue(all(-1e-12<=v<=1+1e-12 for v in r['arms'][arm]['independent_mass']))
        for k,v in d.summarize(saved['records']).items():self.assertEqual(v,saved[k])


if __name__=='__main__':unittest.main()
