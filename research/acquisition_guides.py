"""Post-hoc pre-outcome guide reconstruction with journal forecast parity."""
import gzip
import hashlib
import json
from collections import defaultdict
from pathlib import Path
import sys

PRIOR = [.7, .1, .1, .1]


def predictive(weights):
    if weights is None:
        return PRIOR[:]
    total = sum(weights)
    assert abs(total-1) < 1e-9
    return [.998*w/total + .002*p for w,p in zip(weights, PRIOR)]


def forecast(weights, experts):
    return sum(w*max(1e-6,min(1-1e-6,p)) for w,p in zip(predictive(weights),experts))


def observe(weights, experts, outcome):
    values=[]
    for w,p in zip(predictive(weights),experts):
        p=max(1e-6,min(1-1e-6,p))
        values.append(w*(p if outcome else 1-p))
    total=sum(values)
    return [v/total for v in values]


def summarize(path):
    raw=json.loads(gzip.decompress(path.read_bytes()))
    assert len(raw["Records"])==576
    groups=defaultdict(lambda:dict(frames=0, fitted_frames=0, incumbent=0, short=0,
                                 long=0, tree=0, tree_mass=0., inner_tree_wins=0))
    forecasts_checked=0
    for r in raw["Records"]:
        outer=[None]*4
        inner=None
        for t,tick in enumerate(r["Ticks"]):
            fitted=t>0 and r["Ticks"][t-1]["Audits"]>=32
            for arm in range(4):
                ex=tick["Predictions"][arm]["Experts"]
                assert abs(forecast(outer[arm],ex)-tick["Predictions"][arm]["P"])<1e-12
                forecasts_checked+=1
                if arm==2:
                    assert abs(forecast(inner,r["Inner"][t][arm])-ex[1])<1e-12
                w=outer[arm] if outer[arm] is not None else PRIOR
                iw=inner if inner is not None else PRIOR
                guide=0
                if fitted and w[1]>w[guide]:guide=1
                if fitted and w[2]>w[guide]:guide=2
                use_tree=arm==1 or (arm==2 and sum(iw[1:])>iw[0])
                label="tree" if guide==1 and use_tree else ("incumbent","short","long")[guide]
                key=(r["Generator"],r["Scenario"],r["Split"],r["Mode"],arm)
                g=groups[key]
                g["frames"]+=1
                g["fitted_frames"]+=fitted
                g[label]+=1
                if arm==2:
                    g["inner_tree_wins"]+=fitted and sum(iw[1:])>iw[0]
                    g["tree_mass"]+=predictive(outer[arm])[1]*sum(predictive(inner)[1:])
            # Advance only with the actual journaled forecasts in arrival order.
            for origin in tick.get("Delivered") or []:
                assert 0<=origin<=t
                y=r["Ticks"][origin]["Outcome"]
                inner=observe(inner,r["Inner"][origin][2],y)
                for arm in range(4):
                    outer[arm]=observe(outer[arm],r["Ticks"][origin]["Predictions"][arm]["Experts"],y)
    return dict(post_hoc=True,raw_sha256=hashlib.sha256(path.read_bytes()).hexdigest(),
                forecasts_checked=forecasts_checked,
                groups=[dict(zip(("generator","scenario","split","mode","arm"),k),**v) for k,v in sorted(groups.items())])


if __name__=="__main__":
    out=summarize(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as f:json.dump(out,f,indent=2);f.write("\n")
    print("Forecasts checked",out["forecasts_checked"])
    for row in out["groups"]:
        if row["generator"]=="clustered" and row["scenario"]=="shift128" and row["arm"]==2:print(row)
