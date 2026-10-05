import fs from 'node:fs';
import crypto from 'node:crypto';
import {fileURLToPath} from 'node:url';

const cache=new Map();
export function binomialUpper(k,n,alpha) {
  if(!Number.isInteger(n)||n<1||n>4096||!Number.isInteger(k)||k<0||k>n||!Number.isFinite(alpha)||alpha<=0||alpha>=1) throw Error('invalid binomial inputs');
  const key=`${k}/${n}/${alpha}`;
  if(cache.has(key))return cache.get(key);
  let answer;
  if(k===n)answer=1;
  else if(k===0)answer=-Math.expm1(Math.log(alpha)/n);
  else {
    let lo=0,hi=1;
    for(let iteration=0;iteration<80;iteration++) {
      const p=(lo+hi)/2;
      let lp=n*Math.log1p(-p),cdf=Math.exp(lp);
      for(let i=1;i<=k;i++) {
        lp+=Math.log(n-i+1)-Math.log(i)+Math.log(p)-Math.log1p(-p);
        cdf+=Math.exp(lp);
      }
      if(cdf>alpha)lo=p;else hi=p;
    }
    answer=hi;
  }
  cache.set(key,answer);
  return answer;
}

export function pairedUpper(h,b,n,alpha) {
  if(!Number.isFinite(alpha)||alpha<=0||alpha>=1)throw Error('invalid confidence error probability');
  if(!Number.isInteger(h)||!Number.isInteger(b)||h<0||b<0||h+b>n)throw Error('invalid discordances');
  const d=h+b,a=alpha/3;
  const sUpper=binomialUpper(d,n,a),sLower=1-binomialUpper(n-d,n,a);
  const rUpper=d===0?1:binomialUpper(h,d,a);
  const upper=(rUpper>=.5?sUpper:sLower)*(2*rUpper-1);
  return {upper,sLower,sUpper,rUpper};
}

export function analyze() {
  const path='docs/experiments/mmm-evidence-allocation-v72.jsonl';
  const raw=fs.readFileSync(path),[header,...rows]=raw.toString().trim().split('\n').map(JSON.parse);
  if(header.Version!=='v72'||rows.length!==10240)throw Error('wrong source artifact');
  const hash=data=>crypto.createHash('sha256').update(data).digest('hex');
  for(const [p,want]of Object.entries(header.Hashes))if(hash(fs.readFileSync(p))!==want)throw Error(`source changed ${p}`);
  const cells=[];
  for(const split of ['design','confirmation'])for(const scenario of ['homogeneous128','sparse128','sparse256','negative256','sparse384','weak256']) {
    const at=scenario.endsWith('128')?128:scenario.endsWith('384')?384:256;
    const rs=rows.filter(r=>r.Split===split&&r.Scenario===scenario);
    if(rs.length!==512)throw Error('bad cell size');
    const harmful=rs.filter(r=>r.First[2]<at&&r.First[0]>=at).length;
    const beneficial=rs.filter(r=>r.First[0]<at&&r.First[2]>=at).length;
    const originalUpper=binomialUpper(harmful,512,.05/12);
    cells.push({split,scenario,n:512,harmful,beneficial,excess:(harmful-beneficial)/512,originalUpper,...pairedUpper(harmful,beneficial,512,.05/12)});
  }
  const sources={};
  for(const p of ['research/paired-risk-v73.mjs','research/paired-risk-v73.test.mjs','docs/experiments/mmm-paired-risk-v73-protocol.md'])sources[p]=hash(fs.readFileSync(p));
  return {version:'v73',scope:'post-hoc diagnostic, not fresh confirmation',sourceArtifact:path,sourceArtifactSHA256:hash(raw),sources,cells};
}

if(process.argv[1]===fileURLToPath(import.meta.url)) {
  const result=analyze();
  if(process.argv[2])fs.writeFileSync(process.argv[2],JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});
  console.log(JSON.stringify(result,null,2));
}
