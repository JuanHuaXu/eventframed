import copy
import hashlib
import json
from pathlib import Path
import unittest
import sparse_confirmation as s


class ConfirmationTests(unittest.TestCase):
    def test_replay_and_no_confirmation_label_influence(self):
        root=Path(__file__).parent
        saved=json.loads((root/'sparse-confirmation-v1-results.json').read_text())
        for p,h in saved['sha256'].items():
            self.assertEqual(hashlib.sha256((root/p).read_bytes()).hexdigest(),h)
        design=json.loads((root/'design-preflight-v1.json').read_text())
        confirm=json.loads((root/'confirmation-preflight-v1.json').read_text())
        qs={q['case_id']:q for q in json.loads((root/'prepared-v2/queries.json').read_text())}
        oracle=json.loads((root/'prepared-v2/oracle.json').read_text())
        training={k:v for k,v in oracle.items() if qs[k]['split']=='design'}
        evaluation={k:v for k,v in oracle.items() if qs[k]['split']=='confirmation'}
        result=s.run(design,confirm,qs,training,evaluation)
        self.assertEqual(result['metrics'],saved['metrics'])
        self.assertEqual(result['predictions'],saved['predictions'])
        changed=copy.deepcopy(evaluation)
        for v in changed.values():
            if v['support']:v['support']=['m10-venus']
        alternate=s.run(design,confirm,qs,training,changed)
        for a,b in zip(result['predictions'],alternate['predictions']):
            self.assertEqual([c['learned'] for c in a['candidates']],
                             [c['learned'] for c in b['candidates']])


if __name__=='__main__':unittest.main()
