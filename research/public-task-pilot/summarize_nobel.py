import hashlib
import json
from pathlib import Path

root=Path(__file__).parent
oracle=json.loads((root/'nobel-v1/oracle.json').read_text())
report={};files=[]
for arm in ['baseline','lexical','sparse']:
    path=root/f'nobel-v1/{arm}-results.json';files.append(path)
    rows=json.loads(path.read_text());report[arm]={}
    for kind in ['literal','paraphrase','absent']:
        selected=[r for r in rows if oracle[r['case']]['wording']==kind]
        if arm!='baseline':assert all(r['callback_frontier']==19 for r in selected)
        report[arm][kind]=dict(n=len(selected),top1=sum(r['support_rank']==1 for r in selected),
            survived=sum(r['support_rank']>0 for r in selected),
            mrr=sum(1/r['support_rank'] if r['support_rank'] else 0 for r in selected)/len(selected),
            top_ties=[r['case'] for r in selected if sum(c['score']==max(x['score'] for x in r['candidates']) for c in r['candidates'])>1],
            max_scores=[max(c['score'] for c in r['candidates']) for r in selected])
passed=all(report['sparse'][kind][m]>=report[control][kind][m]
           for kind in ['literal','paraphrase'] for m in ['top1','survived']
           for control in ['baseline','lexical'])
files.extend(root/p for p in ['nobel-v1/corpus.json','nobel-v1/queries.json','nobel-v1/oracle.json',
    'NOBEL_PROTOCOL.md','prepare_nobel.py','summarize_nobel.py','sparse-go-golden.json'])
repo=root.parent.parent
files.extend(repo/p for p in ['internal/researchsparse/sparse.go','internal/service/research_rank.go','cmd/public-pilot-preflight/main.go'])
result=dict(report=report,screen_passed=passed,sha256={str(p.relative_to(repo)):hashlib.sha256(p.read_bytes()).hexdigest() for p in files})
with (root/'nobel-v1/summary.json').open('x') as f:json.dump(result,f,indent=2)
print(json.dumps(result['report'],indent=2));print('screen_passed:',passed)
