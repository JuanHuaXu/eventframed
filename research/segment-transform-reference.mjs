// Algebraic check for a possible tail-prediction optimization. This does not
// replace the frozen Go segment fitter or measure its performance or quality.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';

let genericChecks=0,booleanChecks=0,maxGenericError=0,maxBooleanError=0;
for(let d=0;d<=9;d++)for(let fixture=0;fixture<2;fixture++){
  const cells=3**d,inputs=2**d;
  const f=Array.from({length:cells},(_,j)=>fixture?((31*j)%29)/29/inputs:Math.sin((j+1)*.71)/10);
  const transformed=f.slice();
  // For each coordinate, add the unknown-coordinate contribution to each
  // known value. After all coordinates, a full cell sums all matching masks.
  for(let stride=1;stride<cells;stride*=3)for(let base=0;base<cells;base+=3*stride)for(let j=0;j<stride;j++){
    transformed[base+j+stride]+=transformed[base+j];
    transformed[base+j+2*stride]+=transformed[base+j];
  }
  const index=(mask,x)=>{let i=0,p=1;for(let bit=0;bit<d;bit++,p*=3)if(mask&(1<<bit))i+=p*(1+((x>>bit)&1));return i;};
  for(let x=0;x<inputs;x++){
    let direct=0;for(let mask=0;mask<inputs;mask++)direct+=f[index(mask,x)];
    const error=Math.abs(direct-transformed[index(inputs-1,x)]);
    assert(error<1e-10);maxGenericError=Math.max(maxGenericError,error);genericChecks++;
  }
  const coefficients=Array.from({length:inputs},(_,j)=>fixture?Math.cos(j*.13)/inputs:((17*j)%13-6)/1024);
  const walsh=coefficients.slice();
  for(let width=1;width<inputs;width*=2)for(let base=0;base<inputs;base+=2*width)for(let j=0;j<width;j++){
    const a=walsh[base+j],b=walsh[base+j+width];walsh[base+j]=a+b;walsh[base+j+width]=a-b;
  }
  for(let x=0;x<inputs;x++){
    let direct=0;for(let mask=0;mask<inputs;mask++){const odd=(x&mask).toString(2).replaceAll('0','').length%2;direct+=(odd?-1:1)*coefficients[mask];}
    const error=Math.abs(direct-walsh[x]);assert(error<1e-10);maxBooleanError=Math.max(maxBooleanError,error);booleanChecks++;
  }
}
assert.equal(genericChecks,2046);assert.equal(booleanChecks,2046);
process.stdout.write(JSON.stringify({scope:'tail-transform algebra only; no quality or performance claim',sourceSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),genericChecks,booleanChecks,maxGenericError,maxBooleanError},null,2)+'\n');
