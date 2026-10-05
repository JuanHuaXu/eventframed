"""Streaming v14 audit with descriptive fitting-group bootstrap intervals."""
import gzip
import hashlib
import json
import math
import random
from collections import Counter,defaultdict
from pathlib import Path
import sys


def bootstrap_patterns():
    rng=random.Random(2026107299)
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
        assert header["ExpectedRecords"]==1152
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
                    view=r["Views"][t][a];assert view["cost"]<=6
                    for s in view["trace"]:assert s["values"]==r["Inputs"][t]&s["observed"]
            data=(r["Inputs"],[(t["Outcome"],t["Audit"],t["Missing"],t["Delivered"],t["Predictions"][0]) for t in r["Ticks"]])
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
            rows[key]=metrics
    assert count==1152 and len(pair_signatures)==576
    patterns=bootstrap_patterns();comparisons=[];failures=defaultdict(list)
    for group in sorted({k[:4] for k in rows}):
        for arm,name in ((2,"adaptive_retained"),(3,"static_retained")):
            for window in ("Full","Post"):
                values={name:[] for name in ("subset","fixed","forest")}
                for fit in range(6):
                    for kind,mode,a in (("subset","subset",arm),("fixed","subset",0),("forest","forest",arm)):
                        values[kind].append(sum(rows[group+(mode,fit,stream)][window][a] for stream in range(2))/2)
                mean={k:sum(v)/6 for k,v in values.items()}
                c=dict(zip(("family","generator","scenario","split"),group),arm=name,window=window,means=mean,fit_means=values)
                for ref in ("fixed","forest"):
                    diffs=[a-b for a,b in zip(values["subset"],values[ref])]
                    c["harm_"+ref]=sum(diffs)/6
                    c["descriptive_ci_"+ref]=interval(diffs,patterns)
                comparisons.append(c)
                reasons=[]
                for ref in ("fixed","forest"):
                    if c["harm_"+ref]>.01:reasons.append(ref+"_harm")
                if group[2]=="shift128" and window=="Post":
                    if c["harm_fixed"]>-.005:reasons.append("fixed_gain")
                    if group[1]=="clustered" and c["harm_forest"]>-.005:reasons.append("forest_gain")
                for reason in reasons:failures[name].append(dict(reason=reason,**c))
    h=hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda:f.read(1024*1024),b""):h.update(chunk)
    return dict(records=count,frames=count*512,raw_sha256=h.hexdigest(),comparisons=comparisons,
                interval_note="descriptive percentile bootstrap over six fitting groups; not simultaneous",
                verdicts={name:dict(passed=not failures[name],failures=failures[name]) for name in ("adaptive_retained","static_retained")})

if __name__=="__main__":
    result=summarize(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as f:json.dump(result,f,indent=2);f.write("\n")
    print({k:dict(passed=v["passed"],failures=len(v["failures"])) for k,v in result["verdicts"].items()})
