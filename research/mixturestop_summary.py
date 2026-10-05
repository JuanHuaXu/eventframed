"""Streaming v19 audit with descriptive fitting-group bootstrap intervals."""
import gzip
import hashlib
import json
import math
import random
from collections import Counter,defaultdict
from pathlib import Path
import sys


def bootstrap_patterns():
    rng=random.Random(2026111299)
    patterns=Counter()
    for _ in range(20000):
        counts=[0]*6
        for _ in range(6):counts[rng.randrange(6)]+=1
        patterns[tuple(counts)]+=1
    return patterns


def interval(values,patterns):
    assert len(values)==6
    dist=sorted((sum(c*v for c,v in zip(counts,values))/6,n) for counts,n in patterns.items())
    out=[]
    for q in (.025,.975):
        target=math.ceil(q*20000);n=0
        for v,count in dist:
            n+=count
            if n>=target:out.append(v);break
    return out


def summarize(path):
    rows={};pair_signatures={};count=0
    with gzip.open(path,"rt") as f:
        header=json.loads(next(f))
        assert header["ExpectedRecords"]==1728
        for p,source in header["Sources"].items():
            assert hashlib.sha256(source.encode()).hexdigest()==header["Hashes"][p]
        for line in f:
            r=json.loads(line);count+=1
            key=(r["Family"],r["Generator"],r["Scenario"],r["Split"],r["Mode"],r["Fit"],r["Stream"])
            assert key not in rows
            assert len(r["Ticks"])==len(r["Inputs"])==len(r["Views"])==512
            for t,tick in enumerate(r["Ticks"]):
                delay=16 if r["Scenario"]=="delayed_missing" else 0
                for origin in tick.get("Delivered") or []:
                    assert 0<=origin<=t-delay and not r["Ticks"][origin]["Missing"]
                for a in range(4):
                    assert 0<tick["Predictions"][a]["P"]<1
                    if r["Mode"]=="mixture_stop" and a==2 and r["Views"][t][a]["stop"]=="confidence":
                        assert tick["Predictions"][a]["P"]<=.1 or tick["Predictions"][a]["P"]>=.9
                    view=r["Views"][t][a];assert view["cost"]<=6
                    for s in view["trace"]:assert s["values"]==r["Inputs"][t]&s["observed"]
            data=(r["Inputs"],[(t["Outcome"],t["Audit"],t["Missing"],t["Delivered"],[t["Predictions"][a] for a in (0,1,3)]) for t in r["Ticks"]])
            digest=hashlib.sha256(json.dumps(data,sort_keys=True).encode()).hexdigest()
            pair=key[:4]+key[5:]
            if pair in pair_signatures:assert pair_signatures[pair]==digest
            else:pair_signatures[pair]=digest
            start=128 if r["Scenario"] in ("shift128","recurring") else 256
            metrics={}
            for window,begin in (("Full",0),("Post",start)):
                ticks=r["Ticks"][begin:]
                metrics[window]=[]
                for a in range(4):
                    b=sum((v["Predictions"][a]["P"]-v["Outcome"])**2 for v in ticks)/len(ticks)
                    acc=sum((v["Predictions"][a]["P"]>=.5)==v["Outcome"] for v in ticks)/len(ticks)
                    assert abs(b-r[window][a]["Brier"])<1e-12 and abs(acc-r[window][a]["Accuracy"])<1e-12
                    metrics[window].append(b)
            metrics["FullCost"]=[sum(v[a]["cost"] for v in r["Views"])/512 for a in range(4)]
            metrics["PostCost"]=[sum(v[a]["cost"] for v in r["Views"][start:])/(512-start) for a in range(4)]
            rows[key]=metrics
    assert count==1728 and len(pair_signatures)==576

    patterns=bootstrap_patterns();comparisons=[];failures=defaultdict(list)
    for group in sorted({k[:4] for k in rows}):
        for bank in ("full_budget","mixture_stop"):
            for window in ("Full","Post"):
                values={name:[] for name in ("candidate","fixed","baseline")}
                for fit in range(6):
                    for kind,mode,a in (("candidate",bank,2),("fixed","current",0),("baseline","current",2)):
                        values[kind].append(sum(rows[group+(mode,fit,stream)][window][a] for stream in range(2))/2)
                c=dict(zip(("family","generator","scenario","split"),group),variant=bank,window=window,
                       means={k:sum(v)/6 for k,v in values.items()},fit_means=values)
                for ref in ("fixed","baseline"):
                    diffs=[a-b for a,b in zip(values["candidate"],values[ref])]
                    c["harm_"+ref]=sum(diffs)/6
                    c["descriptive_ci_"+ref]=interval(diffs,patterns)
                c["mean_cost"]={kind:sum(rows[group+(mode,fit,stream)][window+"Cost"][a] for fit in range(6) for stream in range(2))/12 for kind,mode,a in (("candidate",bank,2),("baseline","current",2),("fixed","current",0))}
                comparisons.append(c)
                reasons=[]
                for ref in ("fixed","baseline"):
                    if c["harm_"+ref]>.01:reasons.append(ref+"_harm")
                if group[2]=="shift128" and window=="Post":
                    if c["harm_fixed"]>-.005:reasons.append("fixed_gain")
                    if group[0]=="majority" and group[1]=="clustered" and c["harm_baseline"]>-.005:reasons.append("baseline_gain")
                if group[0]=="majority" and group[1]=="clustered" and window=="Post" and c["mean_cost"]["candidate"]>4:reasons.append("cost")
                if bank=="mixture_stop":
                    for reason in reasons:failures[bank].append(dict(reason=reason,**c))
    h=hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda:f.read(1024*1024),b""):h.update(chunk)
    return dict(records=count,frames=count*512,raw_sha256=h.hexdigest(),comparisons=comparisons,
                interval_note="descriptive six-group bootstrap; not simultaneous",
                verdicts={name:dict(passed=not failures[name],failures=failures[name]) for name in ("mixture_stop",)})

if __name__=="__main__":
    result=summarize(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as f:json.dump(result,f,indent=2);f.write("\n")
    print({k:dict(passed=v["passed"],failures=len(v["failures"])) for k,v in result["verdicts"].items()})
