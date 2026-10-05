"""One frozen fit; no confirmation feedback. See confirmation protocol."""
import hashlib
import json
from pathlib import Path
import sparse_v1 as s


def run(design, confirmation, queries, design_oracle, evaluation_oracle):
    samples = [(s.features(queries[r['case']]['question'], c['frame']),
                int(c['fixture'] in design_oracle[r['case']]['support']))
               for r in design if design_oracle[r['case']]['support']
               for c in r['candidates']]
    w = s.fit(samples)
    predictions = []
    # No evaluation oracle access until every prediction has been computed.
    for r in confirmation:
        cs = []
        for c in r['candidates']:
            x = s.features(queries[r['case']]['question'], c['frame'])
            p = s.predict(w, x)
            cs.append(dict(fixture=c['fixture'], baseline=c['score'],
                           lexical=x[7], learned=p, blend=.5*(p+c['score'])))
        predictions.append(dict(case=r['case'], candidates=cs))
    for r in predictions:
        support = evaluation_oracle[r['case']]['support']
        r['positive_task'] = bool(support)
        for c in r['candidates']:
            c['target'] = int(c['fixture'] in support)
    positive = [r for r in predictions if r['positive_task']]
    absent = [r for r in predictions if not r['positive_task']]
    metrics = {}
    for arm in ('baseline', 'lexical', 'learned', 'blend'):
        ranks = [next((i+1 for i,c in enumerate(sorted(r['candidates'], key=lambda c:-c[arm])) if c['target']), 0) for r in positive]
        metrics[arm] = dict(top1=sum(r==1 for r in ranks), queries=len(ranks),
                            missing_support=sum(r==0 for r in ranks),
                            mrr=sum(1/r if r else 0 for r in ranks)/len(ranks),
                            brier=sum((c[arm]-c['target'])**2 for r in positive for c in r['candidates'])/sum(len(r['candidates']) for r in positive),
                            absent_mean_probability=sum(c[arm] for r in absent for c in r['candidates'])/sum(len(r['candidates']) for r in absent))
    m=metrics
    passed=(m['learned']['top1']>=max(m['baseline']['top1'],m['lexical']['top1'])
            and m['learned']['brier']<.09 and m['learned']['absent_mean_probability']<=.1)
    return dict(predictions=predictions,metrics=metrics,screen_passed=passed)


def main():
    root=Path(__file__).parent
    inputs=['design-preflight-v1.json','confirmation-preflight-v1.json',
            'prepared-v2/queries.json','prepared-v2/oracle.json']
    design,confirmation,qs,oracle=[json.loads((root/p).read_text()) for p in inputs]
    queries={q['case_id']:q for q in qs}
    training={k:v for k,v in oracle.items() if queries[k]['split']=='design'}
    evaluation={k:v for k,v in oracle.items() if queries[k]['split']=='confirmation'}
    result=run(design,confirmation,queries,training,evaluation)
    paths=inputs+['sparse_v1.py','sparse_confirmation.py','SPARSE_CONFIRMATION_PROTOCOL.md']
    result['sha256']={p:hashlib.sha256((root/p).read_bytes()).hexdigest() for p in paths}
    with (root/'sparse-confirmation-v1-results.json').open('x') as f:
        json.dump(result,f,indent=2);f.write('\n')
    print(json.dumps(dict(metrics=result['metrics'],screen_passed=result['screen_passed']),indent=2))


if __name__=='__main__':
    main()
