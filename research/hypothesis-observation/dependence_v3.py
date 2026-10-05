"""Factorized exact Bayes over known hypotheses and per-test copy modes."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import random
import statistics
import experiment as old

CASES=['independent05','independent20','copied05','copied20','mixed20']
ARMS=['independent','joint_gini','joint_entropy','joint_random','once']


class Joint:
    def __init__(self, table, independent=False):
        self.table=table
        self.w=[1/16]*16
        self.mode=[[1. if independent else .5]*16 for _ in range(8)]
        self.first=[None]*8

    def row(self,t):
        if self.first[t] is None:return self.table[t][:]
        return [m*q+(1-m)*self.first[t] for m,q in zip(self.mode[t],self.table[t])]

    def observe(self,t,y):
        row=self.row(t)
        updated=old.update(self.w,row,y)
        for h in range(16):
            denominator=row[h] if y else 1-row[h]
            if denominator>0:
                self.mode[t][h]*=(self.table[t][h] if y else 1-self.table[t][h])/denominator
        self.first[t]=y if self.first[t] is None else self.first[t]
        self.w=updated

    def select(self,seen,arm,rng):
        pairs=[(t,s) for t in range(8) for s in range(1 if arm=='once' else 4) if (t,s) not in seen]
        if not pairs:return None
        if arm=='joint_random':return rng.choice(pairs)
        square=lambda w:sum(x*x for x in old.classes(w,'noise05'))
        scores=[]
        for t in range(8):
            row=self.row(t);prob=sum(w*q for w,q in zip(self.w,row))
            if arm=='joint_entropy':score=old.entropy([prob,1-prob])
            else:
                score=-square(self.w)
                for y,p in [(False,1-prob),(True,prob)]:
                    if p>0:score+=p*square(old.update(self.w,row,y))
            scores.append(score)
        return max(pairs,key=lambda x:(round(scores[x[0]],14),-x[0],-x[1]))


def episode(case,seed):
    truth=random.Random(seed*10).randrange(16)
    rng=random.Random(seed*10+1)
    tape=[[rng.random() for _ in range(8)] for _ in range(4)]
    table=old.likelihood('noise20' if case.endswith('20') else 'noise05')
    result=dict(case=case,seed=seed,truth=truth,arms={})
    for arm in ARMS:
        model=Joint(table,arm in ('independent','once'));seen=set();trace=[]
        rng=random.Random(seed*10+2)
        for _ in range(16):
            forecast=old.classes(model.w,'noise05')
            pair=model.select(seen,arm,rng);y=None
            if pair is not None:
                t,s=pair
                copied=case.startswith('copied') or (case=='mixed20' and t%2==0)
                y=tape[0 if copied else s][t]<table[t][truth]
                model.observe(t,y);seen.add(pair)
            trace.append(dict(forecast=forecast,pair=pair,outcome=y))
        final=old.classes(model.w,'noise05');chosen=max(range(4),key=final.__getitem__)
        result['arms'][arm]=dict(trace=trace,final=final,
            independent_mass=[sum(w*m for w,m in zip(model.w,v)) for v in model.mode],
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
                summaries.append(dict(split=split,case=case,arm=arm,n=len(rows),**{m:statistics.mean(r['arms'][arm][m] for r in rows) for m in ('curve_brier','final_brier','correct','confident_wrong','queries')}))
            for control in ['independent','joint_random','joint_entropy','once']:
                for metric in ['final_brier','curve_brier']:
                    ds=[r['arms'][control][metric]-r['arms']['joint_gini'][metric] for r in rows]
                    gain=statistics.mean(ds);rad=3.3*statistics.stdev(ds)/math.sqrt(len(ds))
                    comparisons.append(dict(split=split,case=case,control=control,metric=metric,gain=gain,lower=gain-rad,upper=gain+rad))
                    if split==1:
                        if case in ('copied05','copied20','mixed20') and control=='independent' and metric=='final_brier':passed=passed and gain>=.02 and gain-rad>0
                        if case.startswith('independent') and control=='independent' and metric=='curve_brier':passed=passed and gain>=-.02
                        if case=='independent20' and control in ('joint_random','joint_entropy') and metric=='curve_brier':passed=passed and gain-rad>0
    return dict(summaries=summaries,comparisons=comparisons,screen_passed=passed)


def main():
    root=Path(__file__).parent;records=[]
    for split in range(2):
        for j,case in enumerate(CASES):
            for i in range(128):
                r=episode(case,2026102201*1000000+split*100000+j*1000+i);r['split']=split;records.append(r)
    result=dict(records=records,**summarize(records),hashes={p:hashlib.sha256((root/p).read_bytes()).hexdigest() for p in ['dependence_v3.py','DEPENDENCE_PROTOCOL.md','experiment.py']})
    with (root/'dependence-v3.json.gz').open('xb') as f,gzip.GzipFile(fileobj=f,mode='wb',mtime=0) as z:z.write(json.dumps(result).encode())
    result.pop('records')
    with (root/'dependence-v3-summary.json').open('x') as f:json.dump(result,f,indent=2)
    print('screen_passed:',result['screen_passed'])


if __name__=='__main__':main()
