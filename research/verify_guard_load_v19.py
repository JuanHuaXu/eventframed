"""Verify the frozen admission diagnostic without treating rejection as success."""
import hashlib
import json
import math
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PATH = ROOT/'docs/experiments/mmm-guard-load-v19.jsonl'


def require(condition, message):
    if not condition:
        raise ValueError(message)


def percentile(values, q):
    return sorted(values)[math.ceil(q*len(values))-1]/1e6 if values else None


def verify():
    rows = [json.loads(line) for line in PATH.read_text().splitlines()]
    header, arms = rows[0], rows[1:]
    require(header['kind'] == 'header' and header['expected_arms'] == 9 and len(arms) == 9, 'incomplete artifact')
    for name, text in header['Sources'].items():
        require(hashlib.sha256(text.encode()).hexdigest() == header['Hashes'][name], 'embedded source hash: '+name)
    require({(r['Trial'], r['Mode']) for r in arms} == {(t, m) for t in range(3) for m in ['off', 'validate', 'durable']}, 'arm coverage')
    summaries = []
    for r in arms:
        require(not r['Errors'], 'runtime errors')
        require(r['Requests'] == 192 and len(r['ReadNS']) == 192 and len(r['WriteNS']) == 96, 'request accounting')
        require(r['Overlap'] > 0, 'no concurrent reader/writer overlap')
        for field in ['ReadNS', 'WriteNS', 'GuardNS', 'AgeNS']:
            require(all(v >= 0 for v in r[field] or []), 'negative timing')
        if r['Mode'] != 'off':
            require(r['Attempts']+r['Dropped'] == 192, 'queue conservation')
            require(r['Attempts'] == r['Accepted']+r['Busy']+r['Stale'], 'attempt conservation')
            require(r['Validated'] == 50*r['Accepted'], 'partial frontier')
            require(len(r['GuardNS'] or []) == r['Attempts'] and len(r['AgeNS'] or []) == r['Accepted'], 'timing accounting')
        if r['Mode'] == 'durable':
            require(r['Admits'] == r['Discards'] == r['Validated'], 'ledger accounting')
        else:
            require(r['Admits'] == r['Discards'] == 0, 'unexpected ledger work')
        off = next(x for x in arms if x['Trial'] == r['Trial'] and x['Mode'] == 'off')
        summaries.append(dict(trial=r['Trial'], mode=r['Mode'], accepted=r['Accepted'],
            attempts=r['Attempts'], busy=r['Busy'], stale=r['Stale'], dropped=r['Dropped'],
            read_p95_ms=percentile(r['ReadNS'], .95), read_p99_ms=percentile(r['ReadNS'], .99),
            read_p99_ratio=percentile(r['ReadNS'], .99)/percentile(off['ReadNS'], .99),
            write_p99_ms=percentile(r['WriteNS'], .99), guard_p95_ms=percentile(r['GuardNS'], .95),
            accepted_age_p95_ms=percentile(r['AgeNS'], .95), candidates=r['Validated'], ledger_admits=r['Admits']))
    return dict(artifact_sha256=hashlib.sha256(PATH.read_bytes()).hexdigest(),
        runtime={k:header[k] for k in ['GoVersion', 'GOOS', 'GOARCH', 'NumCPU', 'GOMAXPROCS']},
        accounting_verified=True, enabled_attempts=sum(r['Attempts'] for r in arms),
        enabled_accepted=sum(r['Accepted'] for r in arms),
        ledger_admissions=sum(r['Admits'] for r in arms), summaries=summaries)


if __name__ == '__main__':
    result = verify()
    with (PATH.parent/'mmm-guard-load-v19-summary.json').open('x') as f:
        json.dump(result, f, indent=2)
    print(json.dumps(result, indent=2))
