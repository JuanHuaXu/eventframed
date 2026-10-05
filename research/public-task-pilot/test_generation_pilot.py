import json
from pathlib import Path
import unittest
from unittest.mock import patch

from generation_pilot import prepare, grade, run


class GenerationTests(unittest.TestCase):
    def test_preparation_and_prompt_boundary(self):
        root = Path(__file__).parent
        prepared = prepare(root)
        self.assertEqual(prepared, prepare(root))
        self.assertEqual(len(prepared['requests']), 56)
        self.assertFalse(any('oracle' in p for p in prepared['hashes']))
        for row in prepared['requests']:
            self.assertEqual([m['role'] for m in row['messages']], ['system', 'user'])
            data = json.loads(row['messages'][1]['content'])
            self.assertEqual(set(data), {'question', 'memory'})
            self.assertNotIn(row['case'], json.dumps(row['messages']))
            self.assertEqual(len(data['memory']), 0 if row['arm'] == 'no_memory' else 10)
            self.assertEqual(set(row['bindings']), {b['id'] for b in data['memory']})

    def test_grader_negative_controls(self):
        oracle = dict(answer='Physics 2023', support=['fact'])
        def row(answer):
            return dict(bindings={'m0': 'fact', 'm1': 'distractor'}, response={'message': {'content': json.dumps(answer)}})
        answer = dict(award='Physics', year=2023, evidence=['m0'])
        self.assertTrue(grade(row(answer), oracle)['grounded'])
        self.assertFalse(grade(row(dict(answer, year=2022)), oracle)['correct'])
        self.assertFalse(grade(row(dict(answer, year=True)), oracle)['valid'])
        self.assertFalse(grade(row(dict(answer, evidence=['m1'])), oracle)['grounded'])
        self.assertFalse(grade(row(dict(answer, evidence=['missing'])), oracle)['grounded'])
        self.assertFalse(grade(row(dict(answer, evidence=[])), oracle)['grounded'])
        abstain = row(dict(award=None, year=None, evidence=[]))
        self.assertFalse(grade(abstain, oracle)['correct'])
        self.assertTrue(grade(abstain, dict(answer='UNKNOWN', support=[]))['correct'])
        self.assertFalse(grade(dict(error='timeout'), oracle)['valid'])
        self.assertFalse(grade(row(dict(answer, explanation='extra')), oracle)['valid'])

    def test_runner_fresh_requests_and_no_oracle(self):
        prepared = prepare(Path(__file__).parent)
        prepared['requests'] = prepared['requests'][:2]
        calls = []
        class Response:
            def __init__(self, data): self.data = json.dumps(data).encode()
            def __enter__(self): return self
            def __exit__(self, *args): pass
            def read(self): return self.data
        def fake(request, **kwargs):
            if isinstance(request, str):
                return Response({'models': [{'name': 'test', 'digest': 'fixture', 'capabilities': ['completion']}]})
            payload = json.loads(request.data)
            calls.append(payload)
            return Response({'message': {'content': '{"award":null,"year":null,"evidence":[]}'}})
        with patch('urllib.request.urlopen', side_effect=fake):
            rows = list(run(prepared, 'test', 'http://127.0.0.1:11434'))
        self.assertEqual(len(rows), 3)
        self.assertEqual(len(calls), 2)
        for request, source in zip(calls, prepared['requests']):
            self.assertEqual(request['messages'], source['messages'])
            self.assertEqual(len(request['messages']), 2)
            self.assertNotIn('bindings', request)
            self.assertNotIn('tools', request)
        with self.assertRaises(ValueError):
            list(run(prepared, 'test', 'https://unapproved.invalid'))


if __name__ == '__main__':
    unittest.main()
