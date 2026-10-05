import fs from 'node:fs';
const dir='research/public-task-pilot/';
function parse(name){const groups=new Map();for(const line of fs.readFileSync(dir+name,'utf8').trim().split('\n')){const r=JSON.parse(line);const m=r.Output?.match(/(BenchmarkPrivateInsert\/\S+)\s+\d+\s+(\d+) ns\/op.*?\s+(\d+) B\/op\s+(\d+) allocs\/op/);if(m){const g=groups.get(m[1])??[];g.push({ns:+m[2],bytes:+m[3],allocs:+m[4]});groups.set(m[1],g)}}if(groups.size!==4)throw Error('case count');for(const g of groups.values())if(g.length!==3)throw Error('repeat count');return groups}
const control=parse('norm-control-cost.jsonl'),candidate=parse('norm-candidate-cost.jsonl');
const median=a=>a.toSorted((a,b)=>a-b)[1];const rows=[];
for(const [name,c]of control){const s=candidate.get(name);if(!s)throw Error('missing case');const cn=median(c.map(x=>x.ns)),sn=median(s.map(x=>x.ns)),cb=median(c.map(x=>x.bytes)),sb=median(s.map(x=>x.bytes));rows.push({name,controlNS:cn,candidateNS:sn,timeRatio:sn/cn,controlBytes:cb,candidateBytes:sb,bytesRatio:sb/cb,pass:sn/cn<=.9&&sb/cb<=1.05})}
const out={rows,pass:rows.every(r=>r.pass)};fs.writeFileSync(dir+'norm-cost-comparison.json',JSON.stringify(out,null,2),{flag:'wx'});console.log(JSON.stringify(out,null,2));
