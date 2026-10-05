"""Post-hoc component losses on the same acquired masks, not alternate policies."""
import gzip
import hashlib
import json
from collections import defaultdict
from pathlib import Path
import sys

def summarize(path):
    raw=json.loads(gzip.decompress(path.read_bytes()))
    groups=defaultdict(list)
    for r in raw["Records"]:
        start=128 if r["Scenario"] in ("shift128","recurring") else 256
        losses=defaultdict(float)
        n=0
        for t,tick in enumerate(r["Ticks"]):
            if t<start or t==0 or r["Ticks"][t-1]["Audits"]<32:continue
            ex=tick["Predictions"][2]["Experts"]
            p=dict(incumbent=ex[0],short=r["Inner"][t][2][0],tree=r["Inner"][t][2][1],long=ex[2],mixture=tick["Predictions"][2]["P"])
            for name,v in p.items():losses[name]+=(v-tick["Outcome"])**2
            n+=1
        assert n>0
        groups[(r["Generator"],r["Scenario"],r["Split"],r["Mode"])].append({name:v/n for name,v in losses.items()})
    out=[]
    for key,rows in sorted(groups.items()):
        assert len(rows)==8
        out.append(dict(zip(("generator","scenario","split","mode"),key),brier={name:sum(r[name] for r in rows)/8 for name in rows[0]},streams=8))
    return dict(raw_sha256=hashlib.sha256(path.read_bytes()).hexdigest(),post_hoc=True,window="post and fitted only",arm="adaptive_retained",groups=out)

if __name__=="__main__":
    result=summarize(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as f:json.dump(result,f,indent=2);f.write("\n")
    for row in result["groups"]:
        if row["generator"]=="clustered" and row["scenario"]=="shift128":print(row)
