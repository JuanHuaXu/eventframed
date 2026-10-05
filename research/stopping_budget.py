"""Post-hoc stop-reason breakdown; known-law information stays offline."""
import gzip
import hashlib
import json
from collections import defaultdict
from pathlib import Path
import sys
from error_budget import conditional,components

def summarize(path):
    groups=defaultdict(lambda:defaultdict(float))
    with gzip.open(path,"rt") as f:
        next(f)
        for line in f:
            r=json.loads(line)
            if r["Mode"]!="subset64" or r["Generator"]!="clustered" or r["Scenario"]!="shift128" or r["Split"]!="confirmation":continue
            for t in range(128,512):
                view=r["Views"][t][2];last=view["trace"][-1]
                g=groups[(r["Family"],view["stop"])]
                q=conditional(r["Family"],r["Generator"],True,last["observed"],last["values"])
                p=r["Ticks"][t]["Predictions"][2]["P"]
                g["frames"]+=1;g["cost_sum"]+=view["cost"]
                g["complete_relevant_fields"]+=(last["observed"]&7)==7
                for name,value in components(p,q).items():g[name+"_sum"]+=value
    h=hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda:f.read(1024*1024),b""):h.update(chunk)
    return dict(post_hoc=True,raw_sha256=h.hexdigest(),groups=[dict(family=k[0],stop=k[1],**v) for k,v in sorted(groups.items())])

if __name__=="__main__":
    out=summarize(Path(sys.argv[1]))
    with Path(sys.argv[2]).open("x") as f:json.dump(out,f,indent=2);f.write("\n")
    for row in out["groups"]:print(row)
