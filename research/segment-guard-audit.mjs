import fs from 'node:fs';
import readline from 'node:readline';
import assert from 'node:assert/strict';
import {checkSegmentReference} from './segment-v120-reference.mjs';
const[source,output]=process.argv.slice(2);assert(source&&output);
let header=true,states=0,rows=0,forecastsPerHead=0,negativeControls=0;
const maximum={evidenceError:0,maxWeightError:0,maxForecastError:0,maxNoChangeError:0};
for await(const line of readline.createInterface({input:fs.createReadStream(source),crlfDelay:Infinity})){
  const r=JSON.parse(line);if(header){assert.equal(r.Cohort,'spike-independent-v1');header=false;continue;}
  if(r.Index!==0)continue;rows++;
  for(const fit of r.Fits.filter(f=>[0,128,224].includes(f.Clock))){
    const a=checkSegmentReference(r,fit,0);states++;forecastsPerHead+=a.forecasts;
    for(const k of Object.keys(maximum))maximum[k]=Math.max(maximum[k],a[k]);
    if(negativeControls===0){
      const bad=structuredClone(fit);bad.SegmentStarts[0][0]+=.01;
      assert.throws(()=>checkSegmentReference(r,bad,0));
      const altered=structuredClone(r);altered.Steps[fit.Clock].P[10]+=.01;
      assert.throws(()=>checkSegmentReference(altered,fit,0));negativeControls=2;
    }
  }
}
assert.equal(rows,84);assert.equal(states,252);assert.equal(forecastsPerHead,8064);
const result={rows,states,forecastsPerHead,negativeControls,maximum,limitations:'Independent direct-probability reconstruction of index0 at clocks0/128/224 in every phase/case/schedule for segment64/static64. Stratified reference, not all fits or a full fresh Go replay.'};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(result,null,2));
