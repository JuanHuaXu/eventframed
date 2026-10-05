import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
const dir='research/hypothesis-observation/';
const raw=readFileSync(dir+'simplex-guard-noise.json'),d=JSON.parse(raw);
for(const [p,h] of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(p)).digest('hex'),h);
assert(d.maxGap<=1e-12);
const ceilings=d.gates.filter(g=>g.worldNoise===.2&&g.kind==='gain').map(g=>({counts:g.counts,achievedGain:g.gain,numericalUpperGain:g.gain+2*d.maxGap+1e-10,required:.005}));
assert.equal(ceilings.length,10);assert(ceilings.every(c=>c.numericalUpperGain<c.required));
const result={scope:'Numerical projection-gap upper bounds for the matched .20 all-genuine world only; not interval-arithmetic formal certificates',reason:'In the matched genuine world proposal=q, and conditional Brier risk differs from squared distance to q by a constant. Projection gap uses half squared distance, hence the factor two. Expectation preserves the global gap bound. Added 1e-10 is a reporting cushion, not a formal rounding-error proof.',inputSHA256:createHash('sha256').update(raw).digest('hex'),maxHalfSquaredProjectionGap:d.maxGap,ceilings};
writeFileSync(dir+'simplex-guard-ceiling.json',JSON.stringify(result,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(result,null,2));

