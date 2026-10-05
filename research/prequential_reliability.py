"""Past-delivered-label diagnostic on the frozen v19 mixture-stop policy."""
from collections import defaultdict,deque
import gzip
import hashlib
import json
from pathlib import Path
import sys

class Reliability:
    def __init__(self):self.history=[deque(maxlen=64),deque(maxlen=64)]
    def status(self,side):
        h=self.history[side]
        if len(h)<32:return "unknown"
        rate=(sum(correct for correct,_ in h)+1)/(len(h)+2)
        nominal=sum(p for _,p in h)/len(h)
        return "warn" if rate<nominal-.03 else "clear"
    def observe(self,side,correct,nominal):self.history[side].append((correct,nominal))

def confidence_record(r,t):
    p=r["Ticks"][t]["Predictions"][2]["P"]
    return r["Views"][t][2]["stop"]=="confidence" and (p<=.1 or p>=.9)

def replay_record(r):
    state=Reliability();records=[];seen=set()
    for t,tick in enumerate(r["Ticks"]):
        if confidence_record(r,t):
            p=tick["Predictions"][2]["P"];side=int(p>=.5)
            # Record the decision before processing ANY outcomes arriving now.
            records.append(dict(step=t,status=state.status(side),error=side!=tick["Outcome"]))
        for origin in tick.get("Delivered") or []:
            assert 0<=origin<=t and origin not in seen and not r["Ticks"][origin]["Missing"]
            seen.add(origin)
            if confidence_record(r,origin):
                p=r["Ticks"][origin]["Predictions"][2]["P"];side=int(p>=.5)
                state.observe(side,side==r["Ticks"][origin]["Outcome"],max(p,1-p))
    return records

def summarize(path):
    groups=defaultdict(lambda:defaultdict(lambda:dict(frames=0,errors=0)));count=0
    with gzip.open(path,"rt") as f:
        header=json.loads(next(f));assert header["ExpectedRecords"]==1728
        for name,source in header["Sources"].items():assert hashlib.sha256(source.encode()).hexdigest()==header["Hashes"][name]
        for line in f:
            r=json.loads(line)
            if r["Mode"]!="mixture_stop":continue
            count+=1
            start=128 if r["Scenario"] in ("shift128","recurring") else 256
            for v in replay_record(r):
                for window in (("Full","Post") if v["step"]>=start else ("Full",)):
                    key=(r["Family"],r["Generator"],r["Scenario"],r["Split"],window)
                    cell=groups[key][v["status"]];cell["frames"]+=1;cell["errors"]+=v["error"]
    assert count==576
    out=[]
    for key,cells in sorted(groups.items()):
        total=sum(v["frames"] for v in cells.values());errors=sum(v["errors"] for v in cells.values())
        w=cells["warn"];rest_n=total-w["frames"];rest_e=errors-w["errors"]
        out.append(dict(zip(("family","generator","scenario","split","window"),key),counts=dict(cells),
                        warning_fraction=w["frames"]/total if total else None,
                        error_capture=w["errors"]/errors if errors else None,
                        warned_error=w["errors"]/w["frames"] if w["frames"] else None,
                        nonwarned_error=rest_e/rest_n if rest_n else None))
    target=next(v for v in out if (v["family"],v["generator"],v["scenario"],v["split"],v["window"])==("majority","clustered","shift128","confirmation","Post"))
    passed=all(target[k] is not None for k in ("warning_fraction","error_capture","warned_error","nonwarned_error")) and target["warning_fraction"]<=.5 and target["error_capture"]>=.5 and target["warned_error"]>=2*target["nonwarned_error"]
    h=hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda:f.read(1024*1024),b""):h.update(chunk)
    return dict(post_hoc=True,raw_sha256=h.hexdigest(),records=count,passed=passed,target=target,groups=out)

if __name__=="__main__":
    out=summarize(Path(sys.argv[1]));out["source"]=Path(__file__).read_text()
    with Path(sys.argv[2]).open("x") as f:json.dump(out,f,indent=2);f.write("\n")
    print(json.dumps(dict(passed=out["passed"],target=out["target"]),indent=2))
