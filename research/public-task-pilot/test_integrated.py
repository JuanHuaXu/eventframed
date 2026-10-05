import hashlib
import json
from pathlib import Path
import unittest
import check_integrated


class IntegratedTests(unittest.TestCase):
    def test_replay_and_hashes(self):
        root=Path(__file__).parent
        saved=json.loads((root/'integrated-summary.json').read_text())
        report,_=check_integrated.check(root)
        self.assertEqual(report,saved['report'])
        for p,h in saved['sha256'].items():
            self.assertEqual(hashlib.sha256((root.parent.parent/p).read_bytes()).hexdigest(),h)
        self.assertEqual(report['confirmation']['sparse']['survived'],5)
        self.assertGreater(report['confirmation']['sparse']['promoted_packed_records'],0)


if __name__=='__main__':unittest.main()
