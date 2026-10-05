import assert from 'node:assert/strict';
import {readFileSync,writeFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {evaluate,restrict,bounds} from './bernstein-risk.mjs';
import {multiplyAffine} from './noise-envelope.mjs';
const dir='research/hypothesis-observation/',raw=readFileSync(dir+'continuous-risk.json'),d=JSON.parse(raw);
let polynomialChecks=0,maxPolynomialError=0;
for(let degree=0;degree<=10;degree++){
  const c=Array(degree+1).fill(0);c[degree]=1;
  for(const [a,b]of [[0,1],[.2,.8],[0,.4],[.6,1],[.5,.5]]){
    const r=restrict(c,a,b),bound=bounds(r);
    assert(bound.lower<=a**degree+1e-12&&bound.upper>=b**degree-1e-12);
    for(const t of [0,.125,.5,.875,1]){
      const expected=(a+(b-a)*t)**degree,error=Math.abs(evaluate(r,t)-expected);
      assert(error<1e-12);maxPolynomialError=Math.max(maxPolynomialError,error);polynomialChecks++;
    }
  }
  let elevated=c;while(elevated.length<11)elevated=multiplyAffine(elevated,1,1);
  for(const t of [0,.1,.5,.9,1]){assert(Math.abs(evaluate(elevated,t)-t**degree)<1e-12);polynomialChecks++;}
}
const archHashes={};let scoreChecks=0,maxScoreError=0;
for(const name of ['risk-budget-evaluation.json','risk-budget-offgrid.json']){
  const bytes=readFileSync(dir+name),data=JSON.parse(bytes);archHashes[name]=createHash('sha256').update(bytes).digest('hex');
  for(const row of data.results)for(const p of ['risk_budget','random','entropy','tie_entropy'])for(const metric of ['finalBrier','areaBrier']){
    const value=evaluate(d.coefficients[row.mask].scores[p][metric],(row.noise-.05)/.3),error=Math.abs(value-row.scores[p][metric]);
    assert(error<1e-11);maxScoreError=Math.max(maxScoreError,error);scoreChecks++;
  }
}
for(const [file,hash]of Object.entries(d.hashes))assert.equal(createHash('sha256').update(readFileSync(dir+file)).digest('hex'),hash);
assert(raw.equals(execFileSync(process.execPath,[dir+'continuous-risk.mjs'],{maxBuffer:64*1024*1024})));
const out={scope:'Known polynomial identities, archived direct-likelihood score parity, frozen hashes and byte-exact replay; not rigorous floating-point error enclosure',polynomialChecks,maxPolynomialError,scoreChecks,maxScoreError,archHashes,replay:true,inputSHA256:createHash('sha256').update(raw).digest('hex'),verifierSHA256:createHash('sha256').update(readFileSync(dir+'verify-continuous-risk.mjs')).digest('hex')};
writeFileSync(dir+'continuous-risk-verification.json',JSON.stringify(out,null,2)+'\n',{flag:'wx',mode:0o600});console.log(JSON.stringify(out,null,2));
