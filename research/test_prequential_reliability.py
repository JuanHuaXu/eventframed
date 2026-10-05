import copy
import unittest
from prequential_reliability import Reliability,replay_record

class ReliabilityTests(unittest.TestCase):
    def test_support_and_cap(self):
        r=Reliability()
        self.assertEqual(r.status(1),"unknown")
        for _ in range(32):r.observe(1,False,.95)
        self.assertEqual(r.status(1),"warn")
        self.assertEqual(r.status(0),"unknown")
        for _ in range(64):r.observe(1,True,.95)
        self.assertEqual(len(r.history[1]),64)
        self.assertEqual(r.status(1),"clear")

    def test_no_future_or_current_leak(self):
        r=dict(Ticks=[dict(Predictions=[{}, {}, {"P":.95}],Outcome=False,Missing=False,Delivered=[i-2] if i>=2 else []) for i in range(40)],Views=[[{}, {}, {"stop":"confidence"}] for _ in range(40)])
        a=replay_record(r)
        self.assertEqual(a[33]["status"],"unknown")
        self.assertEqual(a[34]["status"],"warn")
        changed=copy.deepcopy(r)
        for i in range(20,40):changed["Ticks"][i]["Outcome"]=True
        b=replay_record(changed)
        self.assertEqual([v["status"] for v in a[:23]],[v["status"] for v in b[:23]])

if __name__=="__main__":unittest.main()
