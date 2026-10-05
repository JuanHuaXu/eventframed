"""Independent-source acquisition screen, not a provenance authenticator."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import random
import statistics
import experiment as old

CASES = ['noise05', 'noise20', 'one_source', 'spoofed']
ARMS = ['naive', 'source_gini', 'source_entropy', 'source_random']


def choose(w, table, seen, sources, arm, rng):
    if arm == 'naive':
        return old.select(w, table, 'noise05', 'target_gini', rng), 0
    available = [(t, s) for t in range(8) for s in range(sources) if (t,s) not in seen]
    if not available:
        return None
    if arm == 'source_random':
        return rng.choice(available)
    scores = []
    square = lambda p: sum(v*v for v in old.classes(p, 'noise05'))
    for row in table:
        p = sum(a*b for a,b in zip(w,row))
        if arm == 'source_entropy':
            scores.append(old.entropy([p,1-p]))
        else:
            scores.append((1-p)*square(old.update(w,row,False)) + p*square(old.update(w,row,True))-square(w))
    return max(available, key=lambda x:(round(scores[x[0]],14),-x[0],-x[1]))


def episode(case,seed):
    truth=random.Random(seed*10).randrange(16)
    rng=random.Random(seed*10+1)
    tape=[[rng.random() for _ in range(8)] for _ in range(4)]
    table=old.likelihood('noise20' if case=='noise20' else 'noise05')
    result=dict(case=case,seed=seed,truth=truth,arms={})
    for arm in ARMS:
        w=[1/16]*16;seen=set();trace=[]
        rng=random.Random(seed*10+2)
        for _ in range(16):
            p=old.classes(w,'noise05')
            pair=choose(w,table,seen,1 if case=='one_source' else 4,arm,rng)
            y=None
            if pair is not None:
                t,source=pair
                y=tape[0 if case=='spoofed' else source][t]<table[t][truth]
                w=old.update(w,table[t],y)
                seen.add(pair)
            trace.append(dict(forecast=p,pair=pair,outcome=y))
        final=old.classes(w,'noise05')
        chosen=max(range(4),key=final.__getitem__)
        result['arms'][arm]=dict(trace=trace,final=final,
            curve_brier=statistics.mean(old.brier(x['forecast'],truth%4) for x in trace),
            final_brier=old.brier(final,truth%4),correct=chosen==truth%4,
            confident_wrong=max(final)>=.9 and chosen!=truth%4,
            queries=sum(x['pair'] is not None for x in trace))
    return result


def summarize(records):
    summaries=[];comparisons=[];passed=True
    for split in range(2):
        for case in CASES:
            rows=[r for r in records if r['split']==split and r['case']==case]
            for arm in ARMS:
                summaries.append(dict(split=split,case=case,arm=arm,n=len(rows),**{
                    m:statistics.mean(r['arms'][arm][m] for r in rows)
                    for m in ('curve_brier','final_brier','correct','confident_wrong','queries')}))
            for control,metric in [('naive','final_brier'),('source_random','curve_brier'),('source_entropy','curve_brier')]:
                ds=[r['arms'][control][metric]-r['arms']['source_gini'][metric] for r in rows]
                gain=statistics.mean(ds);rad=3.3*statistics.stdev(ds)/math.sqrt(len(ds))
                comparisons.append(dict(split=split,case=case,control=control,metric=metric,gain=gain,lower=gain-rad,upper=gain+rad))
                if split==1 and case in ('noise05','noise20'):
                    passed=passed and gain>=.02 and gain-rad>0
                if split==1 and case=='one_source' and control=='naive':
                    passed=passed and gain>=0
    return dict(summaries=summaries,comparisons=comparisons,matched_screen=passed)


def main():
    root=Path(__file__).parent
    paths=['provenance_v2.py','PROVENANCE_PROTOCOL.md','experiment.py']
    records=[]
    for split in range(2):
        for j,case in enumerate(CASES):
            for i in range(128):
                r=episode(case,2026101201*1000000+split*100000+j*1000+i)
                r['split']=split;records.append(r)
    result=dict(records=records,**summarize(records),hashes={p:hashlib.sha256((root/p).read_bytes()).hexdigest() for p in paths})
    with (root/'provenance-v2.json.gz').open('xb') as f,gzip.GzipFile(fileobj=f,mode='wb',mtime=0) as z:
        z.write(json.dumps(result).encode())
    result.pop('records')
    with (root/'provenance-v2-summary.json').open('x') as f:json.dump(result,f,indent=2)
    print(json.dumps(result,indent=2))


if __name__=='__main__':main()
