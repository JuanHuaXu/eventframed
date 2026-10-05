"""Audit v13 paired online policies against frozen mean-screen criteria."""
import gzip
import hashlib
import json
import math
from collections import defaultdict
from pathlib import Path
import sys


def summarize(path):
    raw=json.loads(gzip.decompress(path.read_bytes()))
    for name,source in raw["Sources"].items():
        assert hashlib.sha256(source.encode()).hexdigest()==raw["Hashes"][name]
    assert len(raw["Records"])==576
    groups=defaultdict(list)
    pairs=defaultdict(dict)
    for r in raw["Records"]:
        key=(r["Generator"],r["Scenario"],r["Split"],r["Fit"],r["Stream"])
        assert r["Mode"] not in pairs[key]
        pairs[key][r["Mode"]]=r
        assert len(r["Ticks"])==len(r["Views"])==len(r["Inputs"])==512
        for t,tick in enumerate(r["Ticks"]):
            assert all(0<=origin<=t for origin in tick.get("Delivered") or [])
            for a in range(4):
                p=tick["Predictions"][a]["P"]
                assert math.isfinite(p) and 0<p<1
                view=r["Views"][t][a]
                assert view["cost"]<=6
                for step in view["trace"]:
                    assert step["values"] == r["Inputs"][t]&step["observed"]
        start=128 if r["Scenario"] in ("shift128","recurring") else 256
        for window,begin in (("Full",0),("Post",start)):
            for a in range(4):
                ticks=r["Ticks"][begin:]
                b=sum((v["Predictions"][a]["P"]-v["Outcome"])**2 for v in ticks)/len(ticks)
                acc=sum((v["Predictions"][a]["P"]>=.5)==v["Outcome"] for v in ticks)/len(ticks)
                assert abs(b-r[window][a]["Brier"])<1e-12
                assert abs(acc-r[window][a]["Accuracy"])<1e-12
        groups[key[:3]+(r["Mode"],)].append(r)
    for modes in pairs.values():
        u=modes["forest"]
        for mode in ("subset",):
            r=modes[mode]
            assert u["Inputs"]==r["Inputs"]
            for a,b in zip(u["Ticks"],r["Ticks"]):
                assert all(a[k]==b[k] for k in ("Outcome","Audit","Missing","Delivered"))
                assert a["Predictions"][0]==b["Predictions"][0]
    comparisons=[]
    failures=defaultdict(list)
    for (gen,scenario,split,mode),rows in sorted(groups.items()):
        assert len(rows)==8
        controls=groups[(gen,scenario,split,"forest")]
        for window in ("Full","Post"):
            for arm,name in enumerate(("fixed","replacement","adaptive_retained","static_retained")):
                mean=sum(r[window][arm]["Brier"] for r in rows)/8
                fixed=sum(r[window][0]["Brier"] for r in rows)/8
                uniform=sum(r[window][arm]["Brier"] for r in controls)/8
                c=dict(generator=gen,scenario=scenario,split=split,mode=mode,window=window,arm=name,brier=mean,fixed=fixed,uniform=uniform,harm_fixed=mean-fixed,harm_forest=mean-uniform)
                comparisons.append(c)
                if mode!="subset" or arm<2: continue
                reasons=[]
                if mean-fixed>.01: reasons.append("fixed_harm")
                if mean-uniform>.01: reasons.append("forest_harm")
                if scenario=="shift128" and window=="Post":
                    if mean-fixed>-.005: reasons.append("fixed_gain")
                    if gen=="clustered" and mean-uniform>-.005: reasons.append("forest_gain")
                for reason in reasons: failures[name].append(dict(reason=reason,**c))
    return dict(records=576,frames=576*512,sha256=hashlib.sha256(path.read_bytes()).hexdigest(),comparisons=comparisons,
                verdicts={name:dict(passed=not failures[name],failures=failures[name]) for name in ("adaptive_retained","static_retained")})


if __name__=="__main__":
    out=summarize(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as f:json.dump(out,f,indent=2);f.write("\n")
    print({k:dict(passed=v["passed"],failures=len(v["failures"])) for k,v in out["verdicts"].items()})
