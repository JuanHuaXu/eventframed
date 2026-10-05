"""Export public-only parity fixtures; not new evaluation or fitting choices."""
import json
from pathlib import Path
import sparse_v1 as s

root=Path(__file__).parent
design=json.loads((root/'design-preflight-v1.json').read_text())
confirm=json.loads((root/'confirmation-preflight-v1.json').read_text())
qs={q['case_id']:q for q in json.loads((root/'prepared-v2/queries.json').read_text())}
oracle=json.loads((root/'prepared-v2/oracle.json').read_text())
w=s.fit([(s.features(qs[r['case']]['question'],c['frame']),int(c['fixture'] in oracle[r['case']]['support']))
         for r in design if oracle[r['case']]['support'] for c in r['candidates']])
rows=[]
for r in design+confirm:
    for c in r['candidates']:
        query=qs[r['case']]['question']
        x=s.features(query,c['frame'])
        fields={}
        for line in c['frame'].splitlines():
            name,_,value=line.partition(': ')
            if name in s.FIELDS:fields[name]={'value':value}
        assert query.isascii() and all(v['value'].isascii() for v in fields.values())
        rows.append(dict(query=query,event=fields,features=[x.get(k,0.) for k in range(s.DIM)],probability=s.predict(w,x)))
with (root/'sparse-go-golden.json').open('x') as f:
    json.dump(dict(weights=w,rows=rows),f,indent=2);f.write('\n')
print('exported',len(rows),'public parity rows')
