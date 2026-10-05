"""Post-hoc guide versus emitted-mixture stopping diagnostic on v17."""
import gzip
import hashlib
import json
from collections import defaultdict
from pathlib import Path
import sys
from error_budget import conditional,components


def category(stop,p):
    if stop!="confidence":return "budget_or_other"
    return "guide_confident_mixture_uncertain" if .1<p<.9 else "both_confident"


def summarize(path):
    groups=defaultdict(lambda:defaultdict(float));records=0
    with gzip.open(path,"rt") as f:
        header=json.loads(next(f));assert header["ExpectedRecords"]==1152
        for name,source in header["Sources"].items():assert hashlib.sha256(source.encode()).hexdigest()==header["Hashes"][name]
        for line in f:
            r=json.loads(line)
            if r["Mode"]!="current":continue
            records+=1
            start=128 if r["Scenario"] in ("shift128","recurring") else 256
            for t,tick in enumerate(r["Ticks"]):
                view=r["Views"][t][2];p=tick["Predictions"][2]["P"]
                if view["stop"]=="confidence":assert view["probability"]<=.1 or view["probability"]>=.9
                label=category(view["stop"],p)
                local=(t//128)%2==1 if r["Scenario"]=="recurring" else t>=(512 if r["Scenario"]=="stable05" else 128 if r["Scenario"]=="shift128" else 256)
                last=view["trace"][-1]
                q=conditional(r["Family"],r["Generator"],local,last["observed"],last["values"])
                budget=components(p,q)
                for window in (("Full","Post") if t>=start else ("Full",)):
                    key=(r["Family"],r["Generator"],r["Scenario"],r["Split"],window,label)
                    g=groups[key];g["frames"]+=1;g["cost_sum"]+=view["cost"]
                    g["realized_brier_sum"]+=(p-tick["Outcome"])**2
                    for name,v in budget.items():g[name+"_sum"]+=v
    assert records==576
    h=hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda:f.read(1024*1024),b""):h.update(chunk)
    return dict(post_hoc=True,raw_sha256=h.hexdigest(),records=records,
                groups=[dict(zip(("family","generator","scenario","split","window","category"),k),**v) for k,v in sorted(groups.items())])

if __name__=="__main__":
    out=summarize(Path(sys.argv[1]))
    out["sources"]={p:Path(p).read_text() for p in ("research/stopping_disagreement.py","research/error_budget.py")}
    with Path(sys.argv[2]).open("x") as f:json.dump(out,f,indent=2);f.write("\n")
    for row in out["groups"]:
        if row["generator"]=="clustered" and row["scenario"]=="shift128" and row["split"]=="confirmation" and row["window"]=="Post":print(row)
