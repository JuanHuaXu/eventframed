// Corrupt in memory, never alter the sealed outcomes/checker/source files.
import fs from 'node:fs';
import {spawnSync} from 'node:child_process';
const cases=[['loo','integrated omitted-member row'],['label','distinct available evidence'],['issuedLaw','pre-outcome law'],['finalLaw','final law'],['weights','predictive fit weights'],['packet','packet'],['source','source changed']];
const controls=[];
for(const[mode,expected]of cases){
  const script=`
    import fs from 'node:fs';const original=fs.readFileSync;
    fs.readFileSync=function(path,...args){const raw=original.call(this,path,...args);if(String(path)!=='docs/experiments/mmm-crossfit-v32-design.jsonl')return raw;
      const lines=raw.toString().trim().split('\\n'),manifest=JSON.parse(lines[0]),world=JSON.parse(lines[1]),arm=world.Arms.find(a=>a.Model==='crossfit'&&a.Policy==='stratified_random');
      switch(${JSON.stringify(mode)}){
        case'loo':arm.LOO[0][0]-=.01;break;
        case'label':arm.Trace[0].Useful=!arm.Trace[0].Useful;break;
        case'issuedLaw':arm.Trace[0].Q-=.01;break;
        case'finalLaw':arm.Forecast[0]-=.01;break;
        case'weights':arm.Weights[0]+=.01;break;
        case'packet':arm.Packet[0]=(arm.Packet[0]+1)%150;break;
        case'source':manifest.Sources['internal/researchblend/crossfit.go']='bad';break;
      }
      lines[0]=JSON.stringify(manifest);lines[1]=JSON.stringify(world);return Buffer.from(lines.join('\\n')+'\\n');
    };
    await import('./research/crossfit-v32-verify.mjs');
  `;
  const run=spawnSync(process.execPath,['--input-type=module','-e',script],{encoding:'utf8'});
  if(run.error||run.status===0||!run.stderr.includes(expected))throw Error(`negative ${mode} failed: ${run.error??run.stderr}`);
  controls.push({mode,expected,rejected:true});
}
const result={study:'crossfit-v32',controls,immutableTapeMutation:false};
fs.writeFileSync('docs/experiments/mmm-crossfit-v32-negative-controls.json',JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
