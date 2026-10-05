import json,sys,itertools
counts=[1,1,3,1]
types=[0,1,2,7]
schedule=[(0,t) for t in types]+[(1,t) for t,n in zip(types,counts) for _ in range(n)]
def probability(t,h,noise):
    bits=[(h>>i)&1 for i in range(4)]
    truth=bits[t] if t!=7 else bits[0]^bits[2]
    error=.01 if t==2 else noise
    return (1-error) if truth else error
mass_total=baseline_risk=oracle_risk=0.0
forecast_error=0.0
archived=json.load(sys.stdin)
for index,reports in enumerate(itertools.product([0,1],repeat=10)):
    actual=[]
    local=[]
    for h in range(16):
        weight=1/16
        for (_,t),y in zip(schedule,reports):
            p=probability(t,h,.3)
            weight*=p if y else 1-p
        actual.append(weight)
        w=1/16
        for t in types:
            ys=[y for (_,typ),y in zip(schedule,reports) if typ==t]
            p=probability(t,h,.2)
            root=ys[0]
            fresh=1
            for y in ys[1:]: fresh*=p if y else 1-p
            copied=float(all(y==root for y in ys[1:]))
            w*=(p if root else 1-p)*(.5*fresh+.5*copied)
        local.append(w)
    mass=sum(actual)
    total=sum(local)
    q=[sum(w for h,w in enumerate(actual) if h%4==c)/mass for c in range(4)]
    p=[sum(w for h,w in enumerate(local) if h%4==c)/total for c in range(4)]
    mass_total+=mass
    oracle_risk+=mass*(1-sum(v*v for v in q))
    baseline_risk+=sum(w*sum((p[c]-int(c==h%4))**2 for c in range(4)) for h,w in enumerate(actual))
    forecast_error=max(forecast_error,max(abs(q[c]-archived[index*4+c]) for c in range(4)))
assert abs(mass_total-1)<1e-12
assert forecast_error<1e-12
print(json.dumps(dict(counts=counts,noise=.3,mask=0,mass=mass_total,baseline_brier=baseline_risk,oracle_brier=oracle_risk,maximum_possible_gain=baseline_risk-oracle_risk,required_gain=.005,max_archived_oracle_forecast_error=forecast_error,scope="Independent direct-product Python Bayes-risk calculation; known-noise all-genuine oracle; no protection constraints or correction family restriction")))

