"""Frozen design-only sparse representation screen; see SPARSE_V1_PROTOCOL.md."""
import hashlib
import json
import math
from pathlib import Path
import re

STOPS = set('a an the is was of to at in on and what which when according retained records date s'.split())
FIELDS = ('who', 'what', 'where', 'when', 'why', 'how')
DIM = 264


def tokens(s):
    return set(re.findall(r'[^\W_]+', s.lower())) - STOPS


def features(query, frame):
    if len(query.encode()) > 1024:
        raise ValueError('query too large')
    q = sorted(tokens(query))
    if not q or len(q) > 64:
        raise ValueError('invalid query terms')
    fields = {}
    for line in frame.splitlines():
        name, sep, value = line.partition(': ')
        if name in FIELDS:
            if name in fields or len(value.encode()) > 2048:
                raise ValueError('invalid field')
            fields[name] = tokens(value)
    if set(fields) != set(FIELDS):
        raise ValueError('missing fields')
    x = {0: 1.0}
    union = set().union(*fields.values())
    x[7] = len(set(q) & union) / len(q)
    for j, name in enumerate(FIELDS):
        x[j + 1] = len(set(q) & fields[name]) / len(q)
        for term in q:
            key = f'{name}:{int(term in fields[name])}:{term}'.encode()
            digest = hashlib.sha256(key).digest()
            k = 8 + int.from_bytes(digest[:4], 'big') % 256
            x[k] = x.get(k, 0) + (1 if digest[4] & 1 else -1) / len(q)
    return x


def predict(w, x):
    z = max(-40, min(40, sum(w[k] * v for k, v in x.items())))
    return 1 / (1 + math.exp(-z))


def fit(samples):
    if not samples:
        raise ValueError('empty fit')
    w = [math.log(.1 / .9)] + [0.] * (DIM - 1)
    for _ in range(200):
        g = [0.] * DIM
        for x, y in samples:
            if y not in (0, 1):
                raise ValueError('invalid label')
            d = predict(w, x) - y
            for k, v in x.items():
                g[k] += d * v / len(samples)
        w = [v - .5 * (g[k] + (.01 * v if k else 0)) for k, v in enumerate(w)]
    return w


def evaluate(rows, queries, oracle):
    eligible = [r for r in rows if queries[r['case']]['split'] == 'design'
                and oracle[r['case']]['support']]
    groups = sorted({oracle[r['case']]['cluster'] for r in eligible})
    output = []
    for held in groups:
        train = [r for r in eligible if oracle[r['case']]['cluster'] != held]
        w = fit([(features(queries[r['case']]['question'], c['frame']),
                  int(c['fixture'] in oracle[r['case']]['support']))
                 for r in train for c in r['candidates']])
        for r in eligible:
            if oracle[r['case']]['cluster'] != held:
                continue
            cs = []
            for c in r['candidates']:
                x = features(queries[r['case']]['question'], c['frame'])
                p = predict(w, x)
                cs.append(dict(fixture=c['fixture'], target=int(c['fixture'] in oracle[r['case']]['support']),
                               baseline=c['score'], lexical=x[7], learned=p,
                               blend=.5*(p+c['score']), prior=.1))
            output.append(dict(case=r['case'], held_out=held,
                               training_queries=[x['case'] for x in train], candidates=cs))
    metrics = {}
    for arm in ('baseline', 'lexical', 'learned', 'blend', 'prior'):
        ranks = [next(i+1 for i,c in enumerate(sorted(r['candidates'], key=lambda c: -c[arm]))
                      if c['target']) for r in output]
        brier = sum((c[arm]-c['target'])**2 for r in output for c in r['candidates']) / sum(len(r['candidates']) for r in output)
        metrics[arm] = dict(top1=sum(x==1 for x in ranks), queries=len(ranks),
                            mrr=sum(1/x for x in ranks)/len(ranks), brier=brier)
    return dict(predictions=output, metrics=metrics)


def main():
    root = Path(__file__).parent
    paths = ['design-preflight-v1.json', 'prepared-v2/queries.json',
             'prepared-v2/oracle.json', 'SPARSE_V1_PROTOCOL.md', 'sparse_v1.py']
    rows, qs, oracle = [json.loads((root / p).read_text()) for p in paths[:3]]
    result = evaluate(rows, {q['case_id']: q for q in qs}, oracle)
    result['sha256'] = {p: hashlib.sha256((root/p).read_bytes()).hexdigest() for p in paths}
    m = result['metrics']
    result['screen'] = {a: m[a]['top1'] > m['baseline']['top1'] and
                        m[a]['top1'] >= m['lexical']['top1'] and
                        m[a]['brier'] < m['prior']['brier'] for a in ('learned','blend')}
    with (root / 'sparse-v1-results.json').open('x') as f:
        json.dump(result, f, indent=2)
        f.write('\n')
    print(json.dumps(dict(metrics=m, screen=result['screen']), indent=2))


if __name__ == '__main__':
    main()
