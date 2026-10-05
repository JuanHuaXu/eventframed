"""Known-law diagnostic; deliberately outside the online learner."""
import functools
import gzip
import hashlib
import json
from collections import defaultdict
from pathlib import Path
import sys


@functools.lru_cache(None)
def population(family,generator,local):
    assert family in ("majority","multiplexer")
    assert generator in ("fair","biased","clustered")
    out=[]
    for x in range(512):
        if generator=="fair":w=1/512
        elif generator=="biased":
            w=1.
            for b in range(9):
                p=.8 if b%2 else .2
                w*=p if x&(1<<b) else 1-p
        else:
            a,b=.5,.5
            for bit in range(9):
                if x&(1<<bit):a*=.1;b*=.9
                else:a*=.9;b*=.1
            w=a+b
        offset=0 if local else 6
        a,b,c=(bool(x&(1<<(offset+i))) for i in range(3))
        target=(a+b+c)>=2 if family=="majority" else b if c else a
        out.append((x,w,.95 if target else .05))
    assert abs(sum(w for _,w,_ in out)-1)<1e-12
    return out


@functools.lru_cache(None)
def conditional(family,generator,local,mask,value):
    assert 0<=mask<512 and value&~mask==0
    n,d=0.,0.
    for x,w,p in population(family,generator,local):
        if x&mask==value:n+=w*p;d+=w
    return n/d


def components(p,q):
    noise=.05*.95
    observation=q*(1-q)-noise
    forecast=(p-q)**2
    expected=q*(1-p)**2+(1-q)*p*p
    assert observation>=-1e-12
    assert abs(expected-(noise+observation+forecast))<1e-12
    return dict(noise=noise,observation_deficit=observation,forecast_deficit=forecast,expected_brier=expected)


def summarize(path):
    groups=defaultdict(list);count=0
    with gzip.open(path,"rt") as f:
        header=json.loads(next(f));assert header["ExpectedRecords"]==1728
        for p,s in header["Sources"].items():assert hashlib.sha256(s.encode()).hexdigest()==header["Hashes"][p]
        for line in f:
            r=json.loads(line);count+=1
            assert r["Scenario"] in ("stable05","shift128","recurring","delayed_missing")
            start=128 if r["Scenario"] in ("shift128","recurring") else 256
            sums=defaultdict(lambda:defaultdict(float));ns=defaultdict(int)
            for t,tick in enumerate(r["Ticks"]):
                local=(t//128)%2==1 if r["Scenario"]=="recurring" else t>=(512 if r["Scenario"]=="stable05" else 128 if r["Scenario"]=="shift128" else 256)
                for arm in (0,2):
                    view=r["Views"][t][arm];last=view["trace"][-1]
                    assert last["values"]==r["Inputs"][t]&last["observed"] and view["cost"]<=6
                    q=conditional(r["Family"],r["Generator"],local,last["observed"],last["values"])
                    p=tick["Predictions"][arm]["P"]
                    values=components(p,q)
                    values["realized_brier"]=(p-tick["Outcome"])**2
                    for window in (("Full","Post") if t>=start else ("Full",)):
                        key=(arm,window);ns[key]+=1
                        for name,v in values.items():sums[key][name]+=v
            for (arm,window),values in sums.items():
                key=(r["Family"],r["Generator"],r["Scenario"],r["Split"],r["Mode"],arm,window)
                mean={name:v/ns[(arm,window)] for name,v in values.items()}
                assert abs(mean["realized_brier"]-r[window][arm]["Brier"])<1e-12
                groups[key].append(mean)
    assert count==1728
    out=[]
    for key,rows in sorted(groups.items()):
        assert len(rows)==12
        out.append(dict(zip(("family","generator","scenario","split","mode","arm","window"),key),
                        means={name:sum(r[name] for r in rows)/12 for name in rows[0]},streams=12))
    h=hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda:f.read(1024*1024),b""):h.update(chunk)
    return dict(raw_sha256=h.hexdigest(),post_hoc=True,records=count,groups=out)


if __name__=="__main__":
    result=summarize(Path(sys.argv[1]))
    result["diagnostic_source"]=Path(__file__).read_text()
    with Path(sys.argv[2]).open("x") as f:json.dump(result,f,indent=2);f.write("\n")
    for row in result["groups"]:
        if row["split"]=="confirmation" and row["generator"]=="clustered" and row["scenario"]=="shift128" and row["mode"]=="subset64" and row["arm"]==2 and row["window"]=="Post":print(row)
