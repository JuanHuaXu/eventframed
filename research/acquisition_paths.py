"""Post-hoc path diagnostic; does not alter v11 pass/fail criteria."""
import gzip
import hashlib
import json
from collections import defaultdict
from pathlib import Path
import sys


def paths(path):
    raw=json.loads(gzip.decompress(path.read_bytes()))
    pairs=defaultdict(dict)
    for r in raw["Records"]:
        pairs[(r["Generator"],r["Scenario"],r["Split"],r["Fit"],r["Stream"])][r["Mode"]]=r
    groups=defaultdict(lambda:dict(frames=0,changed_paths=0,changed_final_masks=0,absolute_prediction_change=0.))
    for key,modes in pairs.items():
        a=modes["uniform"]
        for mode in ("empirical","oracle"):
            b=modes[mode]
            for arm in (2,3):
                out=groups[key[:3]+(mode,arm)]
                for t in range(512):
                    x,y=a["Views"][t][arm],b["Views"][t][arm]
                    def path_key(v):
                        return [(s["view"]["scope"],s["view"]["depth"],s["observed"],s["values"]) for s in v["trace"]]
                    out["frames"]+=1
                    out["changed_paths"]+=path_key(x)!=path_key(y)
                    out["changed_final_masks"]+=x["trace"][-1]["observed"]!=y["trace"][-1]["observed"]
                    out["absolute_prediction_change"]+=abs(a["Ticks"][t]["Predictions"][arm]["P"]-b["Ticks"][t]["Predictions"][arm]["P"])
    return dict(raw_sha256=hashlib.sha256(path.read_bytes()).hexdigest(),post_hoc=True,
                groups=[dict(zip(("generator","scenario","split","mode","arm"),k),**v) for k,v in sorted(groups.items())])


if __name__=="__main__":
    result=paths(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as f:json.dump(result,f,indent=2);f.write("\n")
    for row in result["groups"]:
        if row["generator"]=="clustered" and row["scenario"]=="shift128":print(row)
