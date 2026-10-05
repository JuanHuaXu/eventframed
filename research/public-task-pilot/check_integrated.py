"""Audit actual pre-packing outputs without treating tie order as new evidence."""
import hashlib
import json
from pathlib import Path


def check(root):
    oracle=json.loads((root/'prepared-v2/oracle.json').read_text())
    report={};files=[]
    for split in ('design','confirmation'):
        full_path=root/f'integrated-{split}-baseline-full.json';files.append(full_path)
        full={r['case']:r for r in json.loads(full_path.read_text())}
        report[split]={}
        baseline=None
        for arm in ('baseline','lexical','sparse'):
            path=root/f'integrated-{split}-{arm}.json';files.append(path)
            rows=json.loads(path.read_text())
            if arm=='baseline':baseline={r['case']:r for r in rows}
            n=top=survived=promotions=0;ties=[];absent=[]
            for r in rows:
                original={c['fixture']:c for c in full[r['case']]['candidates']}
                assert len(original)==13
                if arm!='baseline':
                    assert r['callback_frontier']==13
                    assert r['packet_certainty']==0
                for c in r['candidates']:
                    ref=original[c['fixture']]
                    assert c['law_bundle']==ref['law_bundle'], ('law changed',arm,r['case'],c['fixture'])
                    assert c['backend_score']==ref['backend_score'], ('backend changed',arm,r['case'],c['fixture'])
                    assert abs(c['score']-ref['score']-c['research_delta'])<1e-12
                base_ids={c['fixture'] for c in baseline[r['case']]['candidates']}
                promotions+=sum(c['fixture'] not in base_ids for c in r['candidates'])
                support=oracle[r['case']]['support']
                if support:
                    n+=1;top+=r['candidates'][0]['fixture'] in support
                    survived+=any(c['fixture'] in support for c in r['candidates'])
                    best=max(c['score'] for c in r['candidates'])
                    tied=[c['fixture'] for c in r['candidates'] if c['score']==best]
                    if len(tied)>1:ties.append(dict(case=r['case'],top_tie=tied,support_in_tie=any(x in support for x in tied)))
                elif oracle[r['case']]['answer']=='UNKNOWN':
                    scores=[c['score'] for c in r['candidates']]
                    absent.append(dict(case=r['case'],mean=sum(scores)/len(scores),maximum=max(scores)))
            report[split][arm]=dict(positive_queries=n,top1=top,survived=survived,
                                   promoted_packed_records=promotions,top_ties=ties,absent=absent)
    return report,files


if __name__=='__main__':
    root=Path(__file__).parent
    report,files=check(root)
    files.extend([Path(__file__),root/'INTEGRATED_PROTOCOL.md'])
    repo=root.parent.parent
    files.extend(repo/p for p in ['cmd/public-pilot-preflight/main.go','internal/researchsparse/sparse.go','internal/service/research_rank.go'])
    result=dict(report=report,sha256={str(p.relative_to(repo)):hashlib.sha256(p.read_bytes()).hexdigest() for p in files})
    with (root/'integrated-summary.json').open('x') as f:json.dump(result,f,indent=2)
    print(json.dumps(report,indent=2))
